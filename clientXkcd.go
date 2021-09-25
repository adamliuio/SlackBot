package main

import (
	"encoding/json"
	"fmt"
)

type XKCDClient struct{}

// https://xkcd.com/2516/info.0.json
// {"month": "9", "num": 2516, "link": "", "year": "2021", "news": "", "safe_title": "Hubble Tension", "transcript": "", "alt": "Oh, wait, I might've had it set to kph instead of mph. But that would make the discrepancy even wider!", "img": "https://imgs.xkcd.com/comics/hubble_tension.png", "title": "Hubble Tension", "day": "15"}

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

func (xk XKCDClient) RetrieveJsonById(id int) (xkj xkcdJSON, err error) {
	var fStr string = "https://xkcd.com/%d/info.0.json"
	var body []byte = utils.RetrieveBytes(fmt.Sprintf(fStr, id), nil)
	err = json.Unmarshal(body, &xkj)
	if err != nil {
		return
	}
	return
}

func (xk XKCDClient) GetStoryById(id int) {
	xk.RetrieveJsonById(id)
}
