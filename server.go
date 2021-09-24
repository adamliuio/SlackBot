package main

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var (
	PORT string = ":8080"
	mw   Middlewares
)

func init() {
	mw = Middlewares{}
}

func server() {
	// http://172.105.117.237:8080/slack/events
	app := fiber.New()
	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("Hello, World 👋!") })
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

	app.Listen(PORT)
}
