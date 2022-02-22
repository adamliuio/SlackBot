package main

import (
	"fmt"
	"testing"
)

// func TestDB(t *testing.T) {
// 	// QueryUserProfileTable()
// 	QueryUserProfileTableWithWhere("1495475781467836416")
// }

func TestAllDB(t *testing.T) {
	db.Init()
	// db.ReturnAllRecords()
	t.Logf("|%s|", db.Query("301155927"))
}

func TestApp(t *testing.T) {
	var lst []string
	var i int = 2
	lst = append(lst, fmt.Sprint(i))
	t.Logf("|%+v\n|", lst)
}

func TestInsertDB(t *testing.T) {
	db.InsertRows([][]string{
		{"30155927", "HackerNews"},
		{"30161626", "HackerNews"},
		{"q1rzzk", "Reddit"},
		{"q1gg29", "Reddit"},
		{"1495475781467836416", "Twitter"},
		{"1495444019265998848", "Twitter"},
	})
}
