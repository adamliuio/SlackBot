package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

type SlackClient struct{}

type SlackMessage struct {
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Color         string `json:"color,omitempty"`
	Fallback      string `json:"fallback,omitempty"`
	CallbackID    string `json:"callback_id,omitempty"`
	ID            int    `json:"id,omitempty"`
	AuthorID      string `json:"author_id,omitempty"`
	AuthorName    string `json:"author_name,omitempty"`
	AuthorSubname string `json:"author_subname,omitempty"`
	AuthorLink    string `json:"author_link,omitempty"`
	AuthorIcon    string `json:"author_icon,omitempty"`
	Title         string `json:"title,omitempty"`
	TitleLink     string `json:"title_link,omitempty"`
	Pretext       string `json:"pretext,omitempty"`
	Text          string `json:"text,omitempty"`
	ImageURL      string `json:"image_url,omitempty"`
	Footer        string `json:"footer,omitempty"`
	FooterIcon    string `json:"footer_icon,omitempty"`
	ThumbURL      string `json:"thumb_url,omitempty"`
	// Fields and actions are not defined.
	MarkdownIn []string    `json:"mrkdwn_in,omitempty"`
	Ts         json.Number `json:"ts,omitempty"` // timestamp
}

type MessageBlocks struct {
	Blocks []MessageBlock `json:"blocks,omitempty"`
	Text   string         `json:"text,omitempty"`
}

type MessageBlock struct {
	Type        string           `json:"type,omitempty"` // section / divider / actions / image
	Text        *ElementText     `json:"text,omitempty"`
	Value       string           `json:"value,omitempty"`
	ActionId    string           `json:"action_id,omitempty"`
	Elements    []*Element       `json:"elements,omitempty"`
	Placeholder *Placeholder     `json:"placeholder,omitempty"`
	ImageUrl    string           `json:"image_url,omitempty"`
	AltText     string           `json:"alt_text,omitempty"`
	Accessory   *Accessory       `json:"accessory,omitempty"`
	Title       *Placeholder     `json:"title,omitempty"`
	Element     *CheckboxElement `json:"element,omitempty"`
	Label       *ElementText     `json:"label,omitempty"`
}

type MessageContextBlock struct {
	Type     string     `json:"type,omitempty"`
	Elements []*Element `json:"elements,omitempty"`
}

type CheckboxElement struct {
	Type     string    `json:"type,omitempty"`
	Options  []Element `json:"options,omitempty"`
	ActionId string    `json:"action_id,omitempty"`
}
type Accessory struct {
	Type     string       `json:"type,omitempty"`
	ImageUrl string       `json:"image_url,omitempty"`
	AltText  string       `json:"alt_text,omitempty"`
	Text     *ElementText `json:"text,omitempty"`
	Value    string       `json:"value,omitempty"`
	ActionId string       `json:"action_id,omitempty"`
}
type Element struct {
	Type      string      `json:"type,omitempty"`
	Text      interface{} `json:"text,omitempty"`
	Value     string      `json:"value,omitempty"`
	ActionId  string      `json:"action_id,omitempty"`
	Emoji     bool        `json:"emoji,omitempty"`
	Image_Url string      `json:"image_url,omitempty"`
	AltText   string      `json:"alt_text,omitempty"`
}

type Placeholder struct {
	Type  string `json:"type,omitempty"`
	Text  string `json:"text,omitempty"`
	Emoji bool   `json:"emoji,omitempty"`
}

type ElementText struct {
	Type  string `json:"type,omitempty"` // plain_text /
	Text  string `json:"text,omitempty"`
	Emoji bool   `json:"emoji,omitempty"`
}

func (sc SlackClient) SendMarkdownText(text, url, imageUrl string) (err error) {
	return sc.sendText(sc.CreateTextBlocks(text, "mrkdwn", imageUrl), url)
}

func (sc SlackClient) SendPlainText(text, url string) (err error) {
	return sc.sendText(sc.CreateTextBlocks(text, "plain_text", ""), url)
}

func (sc SlackClient) CreateTextBlocks(text, textType, imageUrl string) MessageBlocks {
	// textType: "mrkdwn" / "plain_text"
	return MessageBlocks{
		Blocks: []MessageBlock{
			sc.CreateTextBlock(text, textType, imageUrl),
		},
	}
}

func (sc SlackClient) CreateImageBlock(imageUrl, altText string) (mb MessageBlock) {
	mb = MessageBlock{
		Type:     "image",
		ImageUrl: imageUrl,
		AltText:  altText,
	}
	return
}

func (sc SlackClient) CreateTextBlock(text, textType, imageUrl string) (mb MessageBlock) {
	if imageUrl == "" {
		mb = MessageBlock{
			Type: "section",
			Text: &ElementText{
				Type: textType,
				Text: text,
			},
		}
	} else {
		mb = MessageBlock{
			Type: "section",
			Text: &ElementText{
				Type: textType,
				Text: text,
			},
			Accessory: &Accessory{
				Type:     "image",
				ImageUrl: imageUrl,
				AltText:  text,
			},
		}
	}
	return
}

func (sc SlackClient) sendText(msgBlocks MessageBlocks, url string) (err error) { // only supports plain_text & mrkdwn
	return sc.SendBlocks(msgBlocks, url)
}

func (sc SlackClient) SendBlocks(msgBlocks MessageBlocks, url string) (err error) {
	var reqBody []byte
	if flag.Lookup("test.v") == nil { // if this is not in test mode
		reqBody, err = json.Marshal(msgBlocks)
	} else { // if is test mode
		reqBody, err = json.MarshalIndent(msgBlocks, "", "    ")
		log.Println(string(reqBody))
	}
	if err != nil {
		return
	}
	if flag.Lookup("test.v") == nil { // if this is not in test mode
		return sc.SendBytes(reqBody, url, nil)
	} else { // if is test mode
		return sc.SendBytes(reqBody, os.Getenv("SlackWebHookUrlTest"), nil)
	}
}

func (sc SlackClient) DeleteMsg(channelID, ts string) (err error) {
	var url string = "https://slack.com/api/chat.delete"
	var headers = [][]string{{"Authorization", fmt.Sprintf("Bearer %s", os.Getenv("SlackUserOAuthToken"))}}
	var reqBody []byte
	reqBody, _ = json.Marshal(struct {
		Channel string `json:"channel,omitempty"`
		Ts      string `json:"ts,omitempty"`
	}{
		Channel: channelID,
		Ts:      ts,
	})
	return sc.SendBytes(reqBody, url, headers)
}

func (sc SlackClient) SendBytes(reqBody []byte, url string, additionalHeaders [][]string) (err error) {
	var headers = [][]string{
		{"Content-Type", "application/json; charset=utf-8"},
	}
	if len(additionalHeaders) > 0 {
		headers = append(headers, additionalHeaders...)
	}
	var respData []byte
	respData, err = utils.HttpRequest("POST", reqBody, url, headers)

	var slackResponse struct {
		Ok bool `json:"ok,omitempty"`
	}
	json.Unmarshal(respData, &slackResponse)
	if string(respData) != "ok" && !slackResponse.Ok {
		err = fmt.Errorf("non-ok response returned from Slack, message: %s", string(respData))
		return
	}
	return
}
