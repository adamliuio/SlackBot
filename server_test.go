package main

import (
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	os.Setenv("DoNotAutoRetrieve", "yes")
	wg := sync.WaitGroup{}
	wg.Add(1)
	go server()
	go func() {
		time.Sleep(3 * time.Second)
		var body []byte
		var err error
		if body, err = utils.HttpRequest("POST", nil, "http://127.0.0.1:8080/discord/hi", nil); err != nil {
			log.Panic(err)
		}
		t.Log(string(body))
	}()
	wg.Wait()
}
