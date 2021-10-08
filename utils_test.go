package main

import (
	"errors"
	"testing"
)

func TestCaller(t *testing.T) {
	var err error = errors.New("damn")
	utils.DealWithError(err)
}
