package main

import (
	"testing"
)

func TestAllDB(t *testing.T) {
	var savedItems []SavedItem = db.ReturnAllRecords()
	for _, item := range savedItems {
		t.Log(item)
	}
}

func TestUpdateXkcd(t *testing.T) {
	for i := 0; i < 10; i++ {
		var item SavedItem = db.UpdateXkcd()
		t.Logf("item: %+v\n", item)
	}
}
