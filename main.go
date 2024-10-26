package main

import (
    "log"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
    app := fiber.New()

    app.Use(cors.New())

    app.Use(func (c *fiber.Ctx) error {
        if c.Is("json") {
            return c.Next()
        }
        return c.SendString("Only JSON allowed!")
    })

    app.Get("/", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "message": "Hello World",
        })
    })

    log.Fatal(app.Listen(":3000"))
}
