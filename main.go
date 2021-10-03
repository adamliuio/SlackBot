package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	rc RedditClient
	sc SlackClient
	tc TwitterClient
	dc DiscordClient
	xk XKCDClient
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
	Hostname, _ = os.Hostname()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
	}
	go rou.StartAll()
}

func main() {
	server()
}
