package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	sc SlackClient
	rc RedditClient
	// xk XKCDClient
	hn = HNClient{
		ItemUrlTmplt:    "https://hacker-news.firebaseio.com/v0/item/%d.json",                // "https://hacker-news.firebaseio.com/v0/item/8863.json?print=pretty"
		StoriesUrlTmplt: "https://hacker-news.firebaseio.com/v0/%sstories.json?print=pretty", // for finding top/new/best stories
		PageUrlTmplt:    "https://news.ycombinator.com/item?id=%d",                           // link to the HN page of this story
	}
	utils    Utils
	rou      Routines
	Hostname string
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
	}
	sc = SlackClient{
		WebHookUrlHN:   os.Getenv("WebHookUrlHN"),
		WebHookUrlTest: os.Getenv("WebHookUrlTest"),
	}
	rc = RedditClient{
		WebHookUrlReddit: os.Getenv("WebHookUrlReddit"),
	}
	Hostname, _ = os.Hostname()
	// rc.Init()
	go rou.StartAll()
}

func main() {
	server()
}
