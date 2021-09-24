package main

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestRetrieveNew(t *testing.T) {
	hn.RetrieveNew("500")
}

const f string = "ids-reddit.json"

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
	var hn HNItem = utils.GetItemById(hn.ItemUrlTmplt, id)
	var receivedJson string = `{"by":"coldturkey","descendants":824,"id":28621288,"kids":[28621925,28626278,28625052,28628542,28625768,28626517,28622228,28622939,28628614,28622049,28626784,28622998,28622821,28623289,28628105,28623861,28628683,28622602,28622499,28626706,28622616,28622905,28622702,28622389,28628060,28623761,28623193,28624865,28624558,28622483,28627504,28627535,28622505,28622886,28622888,28622946,28621745,28622332,28622447,28627210,28626664,28623744,28621959,28625516,28622096,28626032,28626959,28626623,28626245,28627186,28622062,28625403,28622333,28625679,28623518,28624130,28626043,28623114,28624331,28622214,28624609,28625170,28626476,28624992,28623822,28624498,28627944,28622501,28623541,28627291,28623981,28625257,28622128,28624080,28625247,28622416,28622903,28621908,28622966,28625577,28623988,28623444,28624456,28622042,28621821,28622978,28623532,28622219,28622724,28622325,28625693,28625565,28625062,28624104,28624064,28623810,28625903,28623395,28622212,28623207,28623122,28627149,28622154,28622201,28621939,28625650,28624958,28628518,28623997,28624534,28622040,28622243,28622245,28622314,28622094,28622704,28625765],"score":459,"time":1632342266,"title":"Lab-grown meat may never be cost-competitive enough to displace traditional meat","type":"story","url":"https://thecounter.org/lab-grown-cultivated-meat-cost-at-scale/"}`
	_, _ = hn, receivedJson
}

func TestUnixTime(t *testing.T) {
	var unixTs int = 1632342266
	var tm string = time.Unix(int64(unixTs), 0).Format("01-02")
	fmt.Println(tm)
}
