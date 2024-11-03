package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"

	"gametime/internal/utils"
	"gametime/src/datastore"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
	"github.com/segmentio/ksuid"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ApiStart struct {
	DemoRedirect bool       `form:"demoredirect"`
	Clocks       []ApiClock `form:"clocks"`
}

type ApiClock struct {
	Name        string        `form:"name"`
	Increment   time.Duration `form:"increment"`
	InitialTime time.Duration `form:"initialTime"`
}

func (ac ApiClock) toClock() datastore.Clock {
	eventLog := make([]datastore.ClockEvent, 1)
	eventLog[0] = datastore.ClockEvent{
		EventType: datastore.STOP,
		Timestamp: time.Now(),
        RemainingTime: time.Second * ac.InitialTime,
	}

	return datastore.Clock{
		ID:             ksuid.New().String(),
		Name:           ac.Name,
		EventLog:       eventLog,
		Increment:      time.Second * ac.Increment,
        InitialTime: time.Second * ac.InitialTime,
	}
}

var (
	ErrLobbyIsPaused = errors.New("lobby is paused")
)

type channelEvent int

const (
	HEARTBEAT     channelEvent = 10
	LOBBY_UPDATED              = 20
)

type GametimeAPI struct {
	sessionStore   *session.Store
	sseConnections map[string][]chan<- channelEvent
	viewEngine     *html.Engine
	db             *datastore.GametimeDB
}

func RegisterAPI(app *fiber.App, engine *html.Engine, sessionStore *session.Store, dbStore *datastore.GametimeDB) *GametimeAPI {
	gapi := GametimeAPI{

		sessionStore:   sessionStore,
		sseConnections: make(map[string][]chan<- channelEvent),
		viewEngine:     engine,
		db:             dbStore,
	}

	// Register middlewares
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(htmxLocationMiddleware)

	// Register routes
	app.Get("/", gapi.index)
	app.Get("/start", gapi.getStart)
	app.Post("/start", gapi.postStart)
	app.Get("/lobby/:lobbyId/view", gapi.getLobbyViewSelect)
	app.Get("/lobby/:lobbyId/view/:viewId", gapi.getLobbyView)
	app.Post("/clock/press/:clockID", gapi.clockPress)

	app.Get("/sse", gapi.sse)

	return &gapi
}

func (g *GametimeAPI) addSSEConnection(lobbyID string, ch chan<- channelEvent) {
	_, ok := g.sseConnections[lobbyID]
	if !ok {
		g.sseConnections[lobbyID] = make([]chan<- channelEvent, 0)
	}

	g.sseConnections[lobbyID] = append(g.sseConnections[lobbyID], ch)
}

func htmxLocationMiddleware(c *fiber.Ctx) error {
	result := c.Next()

	return result
}

func (g *GametimeAPI) clockPress(c *fiber.Ctx) error {
	clockID := c.Params("clockID")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	lobby, err := g.db.GetLobbyByClock(ctx, clockID)
	if err != nil {
		return err
	}

	if !lobby.State.Running {
		return ErrLobbyIsPaused

	}
	_, clock := lobby.ClockByID(clockID)
	if clock.State() != datastore.RUNNING {
		return datastore.ErrClockIsNotActive
	}

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = g.db.AdvanceLobby(ctx, clockID)
	if err != nil {
		return err
	}

	// TODO: In addition to shipping out to these goroutines, we will need to ship to some event system.
	// This allows us to utilize a load balancer or some form of serverless scaling and not require everyone to use the same node.
	// Could be something like rabbitMQ, or kafka. I wonder if we could use something in mongo, like a collection watch. I dunno the refresh rate
	for _, ch := range g.sseConnections[lobby.ID] {
		ch <- LOBBY_UPDATED
	}
	return nil
}

func (g *GametimeAPI) sse(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	// TODO: Validate view exists
	viewID := c.Query("view")
	view := fmt.Sprintf("pages/lobby/view/%s", viewID)

	lobbyID := c.Query("lobbyID")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	lobby, err := g.db.GetLobby(ctx, lobbyID)
	if err != nil {
		return err
	}

	if !lobby.State.Running {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		g.db.StartLobby(ctx, lobbyID)
	}

	chCtx, chCancel := context.WithCancel(context.Background())
	ch := make(chan channelEvent)
	g.addSSEConnection(lobbyID, ch)

	// Heartbeat for the SSE connection, checks for closed connections.
	// Prevents us from keeping a ton of dead connections on hand
	go func() {
		for {
			select {
			case <-chCtx.Done():
				return
			case <-time.After(time.Second * 30):
				ch <- HEARTBEAT
			}
		}
	}()

	// For observability
	connectionId := ksuid.New()
	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		for {
			select {
			case msg := <-ch:
				if msg == HEARTBEAT {
					fmt.Fprintf(w, "event: heartbeat\n\n")
					err = w.Flush()
					if err != nil {
						// This is a proactive approach to check every once in a while if the connection is dead
						// So we dont need to wait for a flush event, potentially holding on to stale connections
						log.WithFields(log.Fields{
							"err":          err,
							"connectionId": connectionId,
						}).Info("SSE heartbeat failed. closing http connection")
						chCancel()
						break
					}

				} else if msg == LOBBY_UPDATED {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()

					lobby, err := g.db.GetLobby(ctx, lobbyID)
					if err != nil {
						log.WithFields(log.Fields{
							"err":          err,
							"connectionId": connectionId,
						}).Error("error occurred while notifying of lobby update")
						return
					}

					buff := bytes.NewBufferString("")
					if err := g.viewEngine.Render(buff, view, lobby); err != nil {
						log.WithFields(log.Fields{
							"err":          err,
							"connectionId": connectionId,
						}).Error("error occurred while notifying of lobby update")
						return
					}

					result := buff.String()
					result = strings.ReplaceAll(result, "\n", " ")
					fmt.Fprintf(w, "event: lobbyUpdate\ndata: %s\n\n", result)

					err = w.Flush()
					if err != nil {
						// Refreshing page in web browser will establish a new
						// SSE connection, but only (the last) one is alive, so
						// dead connections must be closed here.
						log.WithFields(log.Fields{
							"err":          err,
							"connectionId": connectionId,
						}).Info("error while flushing. closing http connection.")
						break
					}

				}

			}
		}
	}))

	return nil
}

func (g *GametimeAPI) index(c *fiber.Ctx) error {
	session, err := g.sessionStore.Get(c)
	if err != nil {
		return err
	}
	id := session.ID()
	session.Save()

	return c.JSON(fiber.Map{
		"message": id,
	})
}

func (g *GametimeAPI) getStart(c *fiber.Ctx) error {
	return c.Render("pages/start", fiber.Map{}, "layouts/main")
}

func (g *GametimeAPI) postStart(c *fiber.Ctx) error {
	clocks := new(ApiStart)

	if err := c.BodyParser(clocks); err != nil {
		return err
	}

	log.Println(clocks)
	session, err := g.sessionStore.Get(c)
	if err != nil {
		return err
	}

	newLobbyId := ksuid.New().String()

	dbClocks := utils.Map(clocks.Clocks, ApiClock.toClock)

	lobby := datastore.Lobby{
		ID: newLobbyId,
		State: datastore.GameState{
			ActiveClockID: dbClocks[0].ID,
			NextClockID:   dbClocks[1].ID,
			Clocks:        dbClocks,
		},
		Config: datastore.GameConfig{},
	}
	log.Println(lobby)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = g.db.SaveLobby(ctx, lobby)

	session.Set("lobbyId", newLobbyId)
	session.Save()

	newRoute := fmt.Sprintf("/lobby/%s/view", newLobbyId)

	// HTMX Redirect
	c.Set("HX-Location", newRoute)
	return c.SendStatus(204)
}

func (g *GametimeAPI) getLobbyViewSelect(c *fiber.Ctx) error {
	lobbyId := c.Params("lobbyId")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	lobby, err := g.db.GetLobby(ctx, lobbyId)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return c.Redirect("/start")
	} else if err != nil {
		userMessage := fmt.Sprintf("Error occurred while loading the lobby details. error: %s", err.Error())
		return errors.New(userMessage)
	}

	return c.Render("pages/lobby/select", lobby, "layouts/main")
}

func (g *GametimeAPI) getLobbyView(c *fiber.Ctx) error {
	lobbyId := c.Params("lobbyId")
	viewId := c.Params("viewId")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	lobby, err := g.db.GetLobby(ctx, lobbyId)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return c.Redirect("/start")
	} else if err != nil {
		userMessage := fmt.Sprintf("Error occurred while loading the lobby details. error: %s", err.Error())
		return errors.New(userMessage)
	}

	view := fmt.Sprintf("pages/lobby/view/%s", viewId)
	return c.Render(view, lobby, "layouts/main", "layouts/viewcontainer")
}
