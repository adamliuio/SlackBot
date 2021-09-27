package main

import (
	"flag"
	"time"
)

type Routines struct{}

func (rou Routines) StartAll() {
	if flag.Lookup("test.v") == nil { // if this is not in test mode
		for {
			go hn.RetrieveNew() // hacker news
			go rc.RetrieveNew() // reddit
			time.Sleep(time.Hour)
		}
	}
}
