package main

import (
	"testing"
)

func TestUpdateXkcd(t *testing.T) {
	for i := 0; i < 3; i++ {
		var item SavedItem = db.UpdateXkcd()
		t.Logf("item: %+v\n", item)
	}
}

func TestReturnAllRecords(t *testing.T) {
	var items []SavedItem = db.ReturnAllRecords("xkcd")
	for _, item := range items {
		t.Logf("item: %+v\n", item)
	}
}
