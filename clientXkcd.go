package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// XKCD
type XKCDClient struct{}

type xkcdJSON struct {
	Month      string `json:"month,omitempty"`
	Num        int    `json:"num,omitempty"`
	Link       string `json:"link,omitempty"`
	Year       string `json:"year,omitempty"`
	News       string `json:"news,omitempty"`
	SafeTitle  string `json:"safe_title,omitempty"`
	Transcript string `json:"transcript,omitempty"`
	Alt        string `json:"alt,omitempty"`
	Img        string `json:"img,omitempty"`
	Title      string `json:"title,omitempty"`
	Day        string `json:"day,omitempty"`
}

func (xk XKCDClient) AutoRetrieveNew() (err error) {
	var lastID int
	_ = json.Unmarshal(utils.ReadFile(xkcdFilename), &lastID)
	lastID++
	var mbs MessageBlocks
	mbs, err = xk.GetStoryById(fmt.Sprintf("%d", lastID))
	if err != nil {
		return
	}
	err = sc.SendBlocks(mbs, os.Getenv("SlackWebHookUrlCartoons"))
	if err != nil {
		return
	}
	j, _ := json.Marshal(lastID)
	utils.WriteFile(j, xkcdFilename)
	return
}

func (xk XKCDClient) RetrieveJsonById(id string) (xkj xkcdJSON, err error) {
	var fStr string = "https://xkcd.com/%s/info.0.json"
	var body []byte
	if body, err = utils.HttpRequest("GET", nil, fmt.Sprintf(fStr, id), nil); err != nil {
		return
	}
	if err = json.Unmarshal(body, &xkj); err != nil {
		return
	}
	return
}

func (xk XKCDClient) GetStoryById(id string) (mbs MessageBlocks, err error) {
	var xkj xkcdJSON
	xkj, err = xk.RetrieveJsonById(id)
	mbs = MessageBlocks{
		Blocks: []MessageBlock{
			{Type: "divider"},
			{
				Type: "section",
				Text: &ElementText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*%s* [No.%s] <%s-%s-%s>\n%s", xkj.SafeTitle, id, xkj.Year, xkj.Month, xkj.Day, xkj.Transcript),
				},
			},
			{
				Type:     "image",
				ImageUrl: xkj.Img,
				AltText:  xkj.Alt,
			},
		},
	}
	return
}
