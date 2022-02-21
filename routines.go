package main

import (
	"fmt"
	"os"
	"time"
)

type Routines struct{}

func (rou Routines) StartAll() {
	if os.Getenv("DoNotAutoRetrieve") == "yes" {
		return
	}
	if !IsTestMode { // if this is not in test mode
		var i int = 0
		for {
			fmt.Println(time.Now().Format("2006-01-02 15:04:05"), ":", "Auto retrieving new posts... ")
			go hn.AutoRetrieveNew() // hacker news
			// go rc.AutoRetrieveNew() // reddit
			go tc.AutoRetrieveNew() // twitter
			if i%12 == 0 {          // run every 12 hours
				go xk.AutoRetrieveNew() // xkcd
			}
			if i%24 == 0 { // run every 24 hours
				go hn.AutoHNClassic() // hacker news classics
			}
			i++
			time.Sleep(time.Hour)
		}
	}
}
