package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	_ = json.Unmarshal(utils.ReadFile(paramsFilename), &Params)
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
