package main

import (
	"errors"
	"testing"
	"time"
)

func TestCaller(t *testing.T) {
	var err error = errors.New("damn")
	utils.DealWithError(err)
}

func TestStack(t *testing.T) {
	utils.DealWithError(errors.New("something's up"))
}

func TestTimezone(t *testing.T) {
	var loc *time.Location
	var err error
	if loc, err = time.LoadLocation(Params.Timezone); err != nil {
		panic(err)
	}
	t.Log(time.Now().In(loc).Format(Params.TimeFormat))
}
