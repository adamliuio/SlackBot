package main

import (
	"log"
	"os"
	"sort"
	"strconv"
	"testing"

	"github.com/turnage/graw/reddit"
)

func TestGraw(t *testing.T) {
	var harvest reddit.Harvest = rc.Graw()
	sort.Slice(harvest.Posts, func(i, j int) bool {
		return harvest.Posts[i].Ups > harvest.Posts[j].Ups
	})
	for _, p := range harvest.Posts {
		log.Printf("up: %d, score: %d, title: %s\n\n", p.Ups, p.Score, p.Title)
		leastScore, _ := strconv.Atoi(os.Getenv("AutoRedditLeaseScore"))
		if p.Ups > int32(leastScore) {
			log.Printf("%+v\n\n", p)
		}
	}
}

func TestFiling(t *testing.T) {
	utils.DownloadFile("https://v.redd.it/2l3jn69r5jo71/DASH_1080.mp4?source=fallback", "ok.mp4", false)
}

func TestHostname(t *testing.T) {
	hostname, err := os.Hostname()
	if err == nil {
		log.Println("hostname:", hostname)
	}
}
