package main

import (
	"flag"
	"time"
)

type Routines struct{}

func (rou Routines) StartAll() {
	if flag.Lookup("test.v") == nil { // if this is not in test mode
		var i int = 0
		for {
			go hn.AutoRetrieveNew() // hacker news
			go rc.AutoRetrieveNew() // reddit
			go tc.AutoRetrieveNew() // twitter
			if i%12 == 0 {          // run every 12 hours
				go xk.AutoRetrieveNew() // xkcd
			}
			i++
			time.Sleep(time.Hour)
		}
	}
}
