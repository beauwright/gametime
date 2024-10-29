package main

import (
	"gametime/src/datastore"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
)

func main() {
        dbStore, err := datastore.New("mongodb://localhost:27017")
        if err != nil {
            log.Fatal(err)
        }

	templateEngine := html.New("./html/", ".html")
	sessionStore := session.New()
	app := fiber.New(fiber.Config{
		Views: templateEngine,
	})

	RegisterAPI(app, templateEngine, sessionStore, dbStore)

	log.Fatal(app.Listen(":3000"))
}
