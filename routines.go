package main

import (
	"flag"
	"os"
	"time"
)

type Routines struct{}

func (rou Routines) StartAll() {
	if flag.Lookup("test.v") == nil { // if this is not in test mode
		for {
			hn.RetrieveNew(os.Getenv("AutoHNLeaseScore")) // hacker news
			// go rc.RetrieveNew()    // reddit
			time.Sleep(6 * time.Hour)
		}
	}
}
