// curl -i -H "Accept: application/json" -H "Content-Type:application/json" -X POST --data "{\"content\": \"Posted Via Command line\"}" https://discord.com/api/webhooks/893683790134771752/QoCSo-FBHprfKLcs5eWSLL6Otvr2TJF8nRYN25ouVOGYmk1c5mpzdMbQY0uh9QGVeen5
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

type DiscordClient struct{}

type DiscordInteractionResponse struct {
	Type int                            `json:"type,omitempty"`
	Data DiscordInteractionResponseData `json:"data,omitempty"`
}

type DiscordInteractionResponseData struct {
	Content string `json:"content,omitempty"`
}

type DiscordInteractionRequest struct {
	Application_Id string                        `application_id:"content,omitempty"`
	Id             string                        `id:"content,omitempty"`
	Token          string                        `token:"content,omitempty"`
	Type           int                           `type:"content,omitempty"`
	User           DiscordInteractionRequestUser `user:"content,omitempty"`
	Version        int                           `version:"content,omitempty"`
}

type DiscordInteractionRequestUser struct {
	Avatar        string `token:"avatar,omitempty"`
	Discriminator string `token:"discriminator,omitempty"`
	Id            string `token:"id,omitempty"`
	Public_Flags  int    `token:"public_flags,omitempty"`
	Username      string `token:"username,omitempty"`
}

func (dc DiscordClient) Interact(c *fiber.Ctx) (err error) {
	var incomingData []byte = c.Body()
	log.Printf("Discord incomingData raw: %+v\n", string(incomingData))
	var req DiscordInteractionRequest
	json.Unmarshal(incomingData, &req)

	var url string = fmt.Sprintf("https://discord.com/api/v8/interactions/%s/%s/callback", req.Id, req.Token)

	var reqBody, respBody []byte
	if reqBody, err = json.Marshal(DiscordInteractionResponse{
		Type: req.Type,
		Data: DiscordInteractionResponseData{
			Content: "Congrats on sending your command!",
		},
	}); err != nil {
		return err
	}
	var header = [][]string{{"Content-Type", "application/json"}}
	if respBody, err = utils.HttpRequest("POST", reqBody, url, header); err != nil {
		return err
	}
	log.Printf("respBody: %+v\n", string(respBody))

	return c.JSON(DiscordInteractionResponse{Type: req.Type})
}
