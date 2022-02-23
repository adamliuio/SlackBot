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
