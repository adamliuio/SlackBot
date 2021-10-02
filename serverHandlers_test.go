package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	server()
}

func TestShortcuts(t *testing.T) {
	os.Setenv("DoNotAutoRetrieve", "yes")
	go server()
	time.Sleep(3 * time.Second)

	var url string = "http://127.0.0.1:8080/slack/shortcuts"
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	client := &http.Client{}
	_, _ = client.Do(req)

	time.Sleep(3 * time.Second)
}

func TestCommandHn(t *testing.T) {
	msgBlocks, err := hn.RetrieveByCommand("top 10-20")
	if err != nil {
		log.Println(err)
	}
	b, err := json.MarshalIndent(msgBlocks, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v\n", string(b))

	log.Printf("%+v\n", msgBlocks)
	err = sc.SendBlocks(msgBlocks, os.Getenv("SlackWebHookUrlTest"))
	if err != nil {
		log.Fatalln(err)
	}
}

func TestCommandTwitter(t *testing.T) {
	msgBlocks, err := hn.RetrieveByCommand("top 10-20")
	if err != nil {
		log.Println(err)
	}
	b, err := json.MarshalIndent(msgBlocks, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v\n", string(b))

	log.Printf("%+v\n", msgBlocks)
	err = sc.SendBlocks(msgBlocks, os.Getenv("SlackWebHookUrlTest"))
	if err != nil {
		log.Fatalln(err)
	}
}
