package main

import (
	"fmt"
	"log"
	"time"

	"gametime/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/segmentio/ksuid"
)

type GametimeAPI struct {
	sessionStore *session.Store
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

func RegisterAPI(app *fiber.App, sessionStore *session.Store) GametimeAPI {
	gapi := GametimeAPI{
		sessionStore: sessionStore,
	}

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

	return gapi
}

func htmxLocationMiddleware(c *fiber.Ctx) error {
	result := c.Next()

	return result
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

	clockSlice := make([]Clock, 0)

	clockSlice = append(clockSlice,
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
	return c.Render(view, lobby, "layouts/main")
}
