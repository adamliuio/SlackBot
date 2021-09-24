package main

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestTimeTick(t *testing.T) {
	// This will print the time every 5 seconds
	var theTime time.Time
	for theTime = range time.Tick(time.Second * 5) {
		log.Println(theTime.Format("2006-01-02 15:04:05"))
	}
}

func TestTimeUnix(t *testing.T) {
	now := time.Now()
	// log.Println(now.Unix())

	// log.Println(now.Round(0))
	yyyy, mm, dd := now.Date()
	tomorrow := time.Date(yyyy, mm, dd+1, 15, 0, 0, 0, now.Location())
	log.Println(tomorrow)
}

func TestString(t *testing.T) {
	now := "ok"
	log.Println("|" + now + "|")
}

func TestRoutine(t *testing.T) {
	c := make(chan string)

	go countCat(c)

	for i := 0; i < 5; i++ {
		message := <-c
		fmt.Println(message)
	}
}

func countCat(c chan string) {
	for i := 0; i < 5; i++ {
		c <- "Cat"
	}
}
