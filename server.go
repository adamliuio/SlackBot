package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var (
	mw Middlewares
)

func init() {
	mw = Middlewares{}
}

func server() {
	app := fiber.New()
	app.Use(logger.New())

	app.Get("/", mw.Home)
	app.Post("/", func(c *fiber.Ctx) error { return c.SendString("Hello, World 👋!") })
	app.Post("/ping", mw.Ping)

	slack := app.Group("/slack") // /slack
	{
		slack.Post("/commands", mw.Commands)   // /slack/commands
		slack.Post("/shortcuts", mw.Shortcuts) // /slack/shortcuts
		slack.Post("/events", mw.Events)       // /slack/events
	}
	slack.Post("/reddit-redirect", mw.Events) // /slack/events

	if Hostname != "MacBook-Pro.local" {
		app.Use("/file", filesystem.New(filesystem.Config{
			Root:   http.Dir("/tmp"),
			Browse: true,
		}))
	}

	if flag.Lookup("test.v") == nil { // if this is not in test mode
		app.Listen(os.Getenv("ServerListenPort"))
	} else { // if is test mode
		app.Listen(os.Getenv("ServerListenDevPort"))
	}
}
