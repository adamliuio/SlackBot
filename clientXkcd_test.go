package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestXKSend(t *testing.T) {
	var mbs MessageBlocks
	var err error
	mbs, err = xk.GetStoryById("614")
	if err != nil {
		t.Fatal(err)
	}
	err = sc.SendBlocks(mbs, os.Getenv("SlackWebHookUrlTest"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestJsonInt(t *testing.T) {
	var lastID int
	_ = json.Unmarshal(utils.ReadFile(xkcdFilename), &lastID)
	t.Log("M:", fmt.Sprintf("%d", lastID))
	j, _ := json.Marshal(lastID)
	utils.WriteFile(j, xkcdFilename)
}
