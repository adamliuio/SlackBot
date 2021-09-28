package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

type Middlewares struct{}

func (mw Middlewares) Home(c *fiber.Ctx) error {
	log.Println(string(c.Request().URI().QueryString()))
	var incomingData []byte = c.Body()
	log.Printf("c.Body(): %+v\n\n", string(incomingData))
	return c.SendString("Hello, World 👋!")
}

func (mw Middlewares) Shortcuts(c *fiber.Ctx) error {
	log.Println("incoming")

	type T struct {
		Payload string `json:"payload,omitempty"`
	}
	t := new(T)
	if err := c.BodyParser(t); err != nil {
		return err
	}

	var payload SCPayload
	json.Unmarshal([]byte(t.Payload), &payload)
	log.Printf("t: %+v\n\n", t)
	log.Printf("payloadt: %+v\n\n", payload)
	var incomingData []byte = c.Body()
	log.Printf("c.Body(): %+v\n\n", string(incomingData))

	var msgBlocks = MessageBlocks{
		Blocks: []MessageBlock{
			{
				Type: "section",
				Text: &ElementText{
					Type: "mrkdwn",
					Text: "This is a mrkdwn section block :ghost: *this is bold*, and ~this is crossed out~, and <https://google.com|this is a link>",
				},
			},
		},
	}

	_ = msgBlocks

	return c.SendString("pong 👋!")
	// return c.JSON(msgBlocks)
}

func (mw Middlewares) Ping(c *fiber.Ctx) error { return c.SendString("pong 👋!") }

func (mw Middlewares) Events(c *fiber.Ctx) error {
	var incomingData []byte = c.Body()
	log.Printf("c.Body(): %+v\n", string(incomingData))
	var cha = make(map[string]string)
	json.Unmarshal(incomingData, &cha)
	log.Printf("cha: %+v\n", cha)

	return c.SendString(cha["challenge"])
}

func (mw Middlewares) Commands(c *fiber.Ctx) error {
	cmd := new(SlashCommand)
	if err := c.BodyParser(cmd); err != nil {
		return err
	}
	// log.Printf("cmd: %+v\n", cmd)

	switch cmd.Command {
	case "/commands": // use "/commands" to trigger this
		return c.JSON(mw.commandCommands())
	case "/hn":
		return c.JSON(mw.commandHn(cmd))
	case "/twt":
		return c.JSON(mw.commandTwitter(cmd))
	default:
		fmt.Printf("%s.\n", cmd.Command)
	}
	return c.SendString("pong 👋!")
}

func (mw Middlewares) commandCommands() MessageBlocks { // use "/commands" to trigger this
	var cmdStr string = "Your friendly commands reminder:\n📺 */command*: returns all your commands for you to see\n📰 */hn* (/hn top 10-20) returns a list of buttons for retrieving buttons to interact with Hacker News."
	var msgBlocks MessageBlocks = sc.CreateTextBlocks(cmdStr, "mrkdwn", "")
	return msgBlocks
}

func (mw Middlewares) commandHn(cmd *SlashCommand) (msgBlocks MessageBlocks) { // use "/hn top 10" to trigger this
	var err error
	msgBlocks, err = hn.RetrieveByCommand(cmd.Text)
	if err != nil {
		log.Println(err)
	}
	return msgBlocks
}

func (mw Middlewares) commandTwitter(cmd *SlashCommand) (msgBlocks MessageBlocks) { // use "/hn top 10" to trigger this
	var err error
	msgBlocks, err = tc.RetrieveByCommand(cmd.Text)
	if err != nil {
		log.Println(err)
	}
	return msgBlocks
}

// curl -X POST http://127.0.0.1:8080/challenge -d '{"challenge": "accepted"}'
// curl -X POST http://172.105.117.237:8080/ping -d '{"user": "john"}'
