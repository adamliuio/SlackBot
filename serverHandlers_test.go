package main

import (
	"encoding/json"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	server()
}

func TestShortcuts(t *testing.T) {
	go server()
	time.Sleep(3 * time.Second)
	http.Get("http://127.0.0.1:8080/shortcuts")
	time.Sleep(3 * time.Second)
}

func TestCommandHn(t *testing.T) {
	msgBlocks, err := hn.GetHNStories("top 10-20")
	if err != nil {
		log.Println(err)
	}
	b, err := json.MarshalIndent(msgBlocks, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v\n", string(b))

	log.Printf("%+v\n", msgBlocks)
	err = sc.SendBlocks(msgBlocks, sc.WebHookUrlTest)
	if err != nil {
		log.Fatalln(err)
	}
}
