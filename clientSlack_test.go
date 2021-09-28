package main

import (
	"log"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestSendPlainText(t *testing.T) {
	sc.SendPlainText("what's *up* https://api.slack.com/reference/messaging/link-unfurling", os.Getenv("WebHookUrlTest"))
}

func TestSendMarkdownText(t *testing.T) {
	tt := "📺 */command*: returns all your commands for you to see\n📰 */hn* (/hn top 10-20) returns a list of buttons for retrieving buttons to interact with Hacker News."
	sc.SendMarkdownText(tt, os.Getenv("WebHookUrlTest"), "")
}

func TestStringSplit(t *testing.T) {
	t.Logf("%q\n", strings.Split(os.Getenv("AutoRedditSub"), " "))
}

func TestUrl(t *testing.T) {
	u, err := url.Parse("https://siongui.github.io/pali-chanting/zh/archives.html")
	if err != nil {
		log.Fatal(err)
	}
	parts := strings.Split(u.Hostname(), ".")
	log.Printf("u.Hostname(): %+v\n\n", u.Hostname())
	log.Printf("parts: %+v\n\n", parts)
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]
	log.Println(domain)
}
