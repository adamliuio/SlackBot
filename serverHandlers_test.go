package main

import (
	"encoding/json"
	"log"
	"testing"
)

// func TestCommandCommands(t *testing.T) {
// 	var msgBlocks MessageBlocks = mw.CommandCommands()

// 	b, err := json.MarshalIndent(msgBlocks, "", "    ")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Printf("%+v\n", string(b))

// 	err = sc.SendBlocks(msgBlocks, sc.WebHookUrlTest)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	log.Printf("%+v\n", msgBlocks)
// }

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
