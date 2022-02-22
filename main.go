package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = json.Unmarshal(utils.ReadFile(paramsFilename), &Params)
	var err error
	if Hostname, err = os.Hostname(); err != nil {
		log.Fatalln(err)
	}
	if err = godotenv.Load(); err != nil {
		log.Fatalln("Error loading .env file: ", err)
	}
	IsLocal = Hostname == "MacBook-Pro.local"         // checking if the app in local
	IsTestMode = strings.Contains(os.Args[0], "test") // checking if it's in test mode
	db.Init()
}

func main() {
	go rou.StartAll()
	server()
}
