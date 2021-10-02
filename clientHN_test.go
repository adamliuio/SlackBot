package main

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestRetrieveNew(t *testing.T) {
	hn.AutoRetrieveNew()
}

const f string = "ids/ids-reddit.json"

var sl []string

func TestJson(t *testing.T) {
	_ = json.Unmarshal(utils.ReadFile(f), &sl)
	log.Println(sl)
	sl = append(sl, "damn")
	j, _ := json.Marshal(sl)
	utils.WriteFile(j, f)
}

func TestGetHNItemById(t *testing.T) {
	var id int = 28621288
	var hn HNItem = hn.getItemById(hn.ItemUrlTmplt, id)
	log.Printf("%+v\n", hn)
}

func TestUnixTime(t *testing.T) {
	var unixTs int = 1632342266
	var tm string = time.Unix(int64(unixTs), 0).Format("01-02")
	fmt.Println(tm)
}
