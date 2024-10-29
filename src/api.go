package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"gametime/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
	"github.com/segmentio/ksuid"
	"github.com/valyala/fasthttp"
)

type LobbyID string

type GametimeAPI struct {
	sessionStore   *session.Store
	sseConnections map[LobbyID][]chan<- int
	viewEngine     *html.Engine
}

type Clock struct {
	ID            string
	Name          string
	Position      int
	Increment     time.Duration
	TimeRemaining time.Duration
}

type ApiStart struct {
	DemoRedirect bool       `form:"demoredirect"`
	Clocks       []ApiClock `form:"clocks"`
}

type ApiClock struct {
	Name        string        `form:"name"`
	Increment   time.Duration `form:"increment"`
	InitialTime time.Duration `form:"initialTime"`
}

func (ac ApiClock) toClock(position int) Clock {
	return Clock{
		ID:            ksuid.New().String(),
		Name:          ac.Name,
		Position:      position,
		Increment:     time.Second * ac.Increment,
		TimeRemaining: time.Second * ac.InitialTime,
	}
}

type GameState struct {
	ActiveClockID string
	NextClockID   string
	Clocks        []Clock
}

type GameConfig struct{}

type Lobby struct {
	ID     string
	State  GameState
	Config GameConfig
}

var base = make([]Clock, 0)

var clockSlice = append(base,
	Clock{
		ID:            ksuid.New().String(),
		Name:          "Hia",
		Position:      0,
		Increment:     time.Second * 15,
		TimeRemaining: time.Second * 30,
	},
	Clock{
		ID:            ksuid.New().String(),
		Name:          "Fren",
		Position:      1,
		Increment:     time.Second * 30,
		TimeRemaining: time.Second * 600,
	},
)

func RegisterAPI(app *fiber.App, engine *html.Engine, sessionStore *session.Store) GametimeAPI {
	gapi := GametimeAPI{

		sessionStore:   sessionStore,
		sseConnections: make(map[LobbyID][]chan<- int),
		viewEngine:     engine,
	}
	fmt.Println("Uhhhh")

	gapi.sseConnections["lobbyID"] = make([]chan<- int, 0)
	fmt.Println("Yaaaah")

	sessionStore.RegisterType(Lobby{})

	// Register middlewares
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

	return gapi
}

func htmxLocationMiddleware(c *fiber.Ctx) error {
	result := c.Next()

	return result
}

func (g *GametimeAPI) clockPress(c *fiber.Ctx) error {
	// TODO: Filter connections to current lobby, invoking only those
	for _, ch := range g.sseConnections["lobbyID"] {
		ch <- 0
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

	ch := make(chan int)
	// TODO: Dont use constant lobbyID here
	g.sseConnections["lobbyID"] = append(g.sseConnections["lobbyID"], ch)

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		for {
			select {
			case <-ch:
				lobby := Lobby{
					ID: "lobbyID",
					State: GameState{
						ActiveClockID: clockSlice[0].ID,
						NextClockID:   clockSlice[1].ID,
						Clocks:        clockSlice,
					},
					Config: GameConfig{},
				}

				buff := bytes.NewBufferString("")
				if err := g.viewEngine.Render(buff, view, lobby); err != nil {
					fmt.Println(err)
					return
				}

				result := buff.String()
				result = strings.ReplaceAll(result, "\n", " ")
				fmt.Fprintf(w, "event: lobbyUpdate\ndata: %s\n\n", result)

				err := w.Flush()
				if err != nil {
					// Refreshing page in web browser will establish a new
					// SSE connection, but only (the last) one is alive, so
					// dead connections must be closed here.
					fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)

					break
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

	// TODO: Persist lobby in datastore
	dbClocks := utils.MapWithIndex(clocks.Clocks, ApiClock.toClock)

	lobby := Lobby{
		ID: newLobbyId,
		State: GameState{
			ActiveClockID: dbClocks[0].ID,
			NextClockID:   dbClocks[1].ID,
			Clocks:        dbClocks,
		},
		Config: GameConfig{},
	}
	log.Println(lobby)

	session.Set("lobbyId", newLobbyId)
	session.Save()

	if clocks.DemoRedirect == true {
		log.Println("Hia")
		newRoute := fmt.Sprintf("/lobby/%s/view", newLobbyId)
		log.Println(newRoute)

		// HTMX Redirect
		c.Set("HX-Location", newRoute)
		return c.SendStatus(204)
	}

	return c.Render("pages/start", fiber.Map{"Error": newLobbyId})

}

func (g *GametimeAPI) getLobbyViewSelect(c *fiber.Ctx) error {
	lobbyId := c.Params("lobbyId")

	clockSlice := make([]Clock, 0)

	clockSlice = append(clockSlice,
		Clock{
			ID:            ksuid.New().String(),
			Name:          "Hia",
			Position:      0,
			Increment:     time.Second * 15,
			TimeRemaining: time.Second * 300,
		},
		Clock{
			ID:            ksuid.New().String(),
			Name:          "Fren",
			Position:      1,
			Increment:     time.Second * 30,
			TimeRemaining: time.Second * 600,
		},
	)

	// TODO: Load from datastore
	lobby := Lobby{
		ID: lobbyId,
		State: GameState{
			Clocks: clockSlice,
		},
		Config: GameConfig{},
	}
	log.Println(lobby)

	return c.Render("pages/lobby/select", lobby, "layouts/main")
}

func (g *GametimeAPI) getLobbyView(c *fiber.Ctx) error {
	lobbyId := c.Params("lobbyId")
	viewId := c.Params("viewId")

	// TODO: Load from datastore
	lobby := Lobby{
		ID: lobbyId,
		State: GameState{
			ActiveClockID: clockSlice[0].ID,
			NextClockID:   clockSlice[1].ID,
			Clocks:        clockSlice,
		},
		Config: GameConfig{},
	}
	log.Println(lobby)

	view := fmt.Sprintf("pages/lobby/view/%s", viewId)
	return c.Render(view, lobby, "layouts/main", "layouts/viewcontainer")
}
