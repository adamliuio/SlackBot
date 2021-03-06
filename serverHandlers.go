package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Middlewares struct{}

func (mw Middlewares) Home(c *fiber.Ctx) error {
	return c.SendString("Hello, World š!")
}

func (mw Middlewares) Shortcuts(c *fiber.Ctx) (err error) {
	t := new(struct {
		Payload string `json:"payload,omitempty"`
	})
	if err := c.BodyParser(t); err != nil {
		return err
	}

	var payload SCPayload
	json.Unmarshal([]byte(t.Payload), &payload)
	if payload.Actions[0].Action_Id == "delete-todo-button" {
		err = sc.DeleteMsg(payload.Container.Channel_Id, payload.Container.MessageTs)
		if err != nil {
			return
		}
	}
	return
}

func (mw Middlewares) Ping(c *fiber.Ctx) error { return c.SendString("pong š!") }

func (mw Middlewares) Commands(c *fiber.Ctx) error {
	cmd := new(SlashCommand)
	if err := c.BodyParser(cmd); err != nil {
		return err
	}

	switch cmd.Command {
	case "/commands":
		return c.JSON(mw.commandCommands())
	case "/hn":
		return c.JSON(mw.commandHn(cmd)) // "/hn top 10"
	case "/hnsearch":
		return c.JSON(mw.commandHnSearch()) // "/hn top 10"
	case "/hnclassicsnew":
		return c.JSON(mw.commandHnClassicNew()) // "/hnclassicsnew"
	case "/todo":
		return c.JSON(mw.commandToDo(cmd)) // "/todo do something"
	case "/twt":
		return c.JSON(mw.commandTwitter(cmd)) // /twt Makers 5
	case "/xkcd":
		return c.JSON(mw.commandXkcd(cmd)) // "/xkcd 123"
	default:
		fmt.Printf("%s.\n", cmd.Command)
	}
	return c.SendString("pong š!")
}

func (mw Middlewares) commandCommands() (mbs MessageBlocks) {
	var cmdStr string = "Your friendly commands reminder:\nšŗ */command*: returns all your commands for you to see\nš° */hn* (/hn top 10-20) returns a list of buttons for retrieving buttons to interact with Hacker News."
	mbs = sc.CreateTextBlocks(cmdStr, "mrkdwn", "")
	return mbs
}

func (mw Middlewares) commandToDo(cmd *SlashCommand) MessageBlocks {
	var todoStrs []string = strings.Split(cmd.Text, "\n")
	for _, str := range todoStrs {
		var mbs MessageBlocks
		mbs.Blocks = append(mbs.Blocks, MessageBlock{
			Type: "section",
			Text: &ElementText{
				Type: "mrkdwn",
				Text: "*" + str + "*",
			},
			Accessory: &Accessory{
				Type: "button",
				Text: &ElementText{
					Type:  "plain_text",
					Text:  "Done",
					Emoji: true,
				},
				Value:    uuid.New().String(),
				ActionId: "delete-todo-button",
			},
		})
		var err error = sc.SendBlocks(mbs, os.Getenv("SlackWebHookUrlTodo"))
		if err != nil {
			log.Panic(err)
		}
	}
	return sc.CreateTextBlocks("to-do added.", "plain_text", "")
}

func (mw Middlewares) commandHn(cmd *SlashCommand) (mbs MessageBlocks) {
	var err error
	mbs, err = hn.RetrieveByCommand(cmd.Text)
	if err != nil {
		log.Println(err)
	}
	return mbs
}

func (mw Middlewares) commandHnClassicNew() MessageBlocks {
	go hn.AutoHNClassic()
	return sc.CreateTextBlocks("new batch of hn classics on the way", "plain_text", "")
}

func (mw Middlewares) commandHnSearch() MessageBlocks {
	go hn.AutoHNClassic()
	return sc.CreateTextBlocks("<https://hn.algolia.com|HN Algolia>", "mrkdwn", "")
}

func (mw Middlewares) commandTwitter(cmd *SlashCommand) (mbs MessageBlocks) {
	var err error
	mbs, err = tc.RetrieveByCommand(cmd.Text)
	if err != nil {
		log.Println(err)
	}
	return mbs
}

func (mw Middlewares) commandXkcd(cmd *SlashCommand) (mbs MessageBlocks) {
	var err error
	mbs, err = xk.GetStoryById(cmd.Text)
	if err != nil {
		log.Println(err)
	}
	return mbs
}

// curl -X POST http://127.0.0.1:8080/challenge -d '{"challenge": "accepted"}'
// curl -X POST http://172.105.117.237:8080/ping -d '{"user": "john"}'
