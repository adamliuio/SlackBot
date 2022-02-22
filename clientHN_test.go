package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestAutoHNClassic(t *testing.T) {
	hn.AutoHNClassic()
}

func TestParams(t *testing.T) {
	_ = json.Unmarshal(utils.ReadFile(paramsFilename), &Params)
	t.Log(Params)
}

func TestAlgolia(t *testing.T) {
	// var url string = "https://hn.algolia.com/api/v1/search?query=github"
	// var url string = "http://hn.algolia.com/api/v1/search?tags=story&numericFilters=created_at_i>%d,created_at_i<%d" // http://hn.algolia.com/api/v1/search_by_date?tags=story&numericFilters=created_at_i>X,created_at_i<Y

	var date string = Params.LatestHNClassicDate
	t.Log("date:", date)

	var results HNAlgoliaSearchResults
	var err error
	if results, err = hn.RetrieveHNClassic(); err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v\n", results)
}

// --- --- --- --- --- --- --- --- --- --- --- --- --- --- --- --- ---

func TestRetrieveNew(t *testing.T) {
	hn.AutoRetrieveNew()
}

var sl []string

func TestJson(t *testing.T) {
	_ = json.Unmarshal(utils.ReadFile("data-samples/t.json"), &sl)
	t.Log(sl)
	sl = append(sl, "damn")
	j, _ := json.Marshal(sl)
	utils.WriteFile(j, "data-samples/t.json")
}

func TestGetHNItemById(t *testing.T) {
	var id int = 28621288
	var hn HNItem = hn.getItemById(hn.ItemUrlTmplt, fmt.Sprint(id))
	t.Logf("%+v\n", hn)
}

func TestUnixTime(t *testing.T) {
	var unixTs int = 1632342266
	var tm string = time.Unix(int64(unixTs), 0).Format("01-02")
	t.Log(tm)
}

func TestFormatTime(t *testing.T) {
	var layoutISO string = "2006-01-02"
	var date string = "1999-12-31"
	tt, _ := time.Parse(layoutISO, date)
	t.Log(tt) // 1999-12-31 00:00:00 +0000 UTC
	t.Log(tt.Format(layoutISO))
}
