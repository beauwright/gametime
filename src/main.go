package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
)


func main() {
    templateEngine := html.New("./html/", ".html")
    sessionStore := session.New()
    app := fiber.New(fiber.Config{
        Views: templateEngine,
    })

    RegisterAPI(app, sessionStore)

    log.Fatal(app.Listen(":3000"))
}
