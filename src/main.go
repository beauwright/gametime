package main

import (
	"errors"
	"gametime/src/datastore"
    log "github.com/sirupsen/logrus"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
)

func main() {
	dbStore, err := datastore.New("mongodb://localhost:27017")
	if err != nil {
		log.Fatal("unable to connect to mongo", err)
	}

	templateEngine := html.New("./html/", ".html")
	sessionStore := session.New()
	app := fiber.New(fiber.Config{
		Views: templateEngine,
        ErrorHandler: func(ctx *fiber.Ctx, err error) error {
            // Status code defaults to 500
            code := fiber.StatusInternalServerError

            // Retrieve the custom status code if it's a *fiber.Error
            var e *fiber.Error
            if errors.As(err, &e) {
                code = e.Code
            }

            requestId := ctx.Context().ID()

            if code == fiber.StatusInternalServerError{
                log.WithFields(log.Fields{
                    "err": err,
                    "requestId": requestId,
                }).Error("unhandled error caught during request")

            }

            return ctx.Render("pages/error", fiber.Map{
                "StatusCode": code,
                "Message": err.Error(),
                "RequestID": requestId,
            }, "layouts/main")
        },
	})

	RegisterAPI(app, templateEngine, sessionStore, dbStore)

	log.Fatal(app.Listen(":3000"))
}
