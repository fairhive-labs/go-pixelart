package main

import (
	"testing"
	"time"
)

func TestGetFilename(t *testing.T) {
	f := "jsie.png"

	date := "Thu Mar 3 15:04:05 2022"
	ti, err := time.Parse("Mon Jan 2 15:04:05 2006", date)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	got := getFilename(f, ti)
	if got != "jsie_20220303-150405.png" {
		t.Errorf("GetFilename(f, ti) = %s; want jsie_20220303-150405.png", got)
		t.FailNow()
	}
}
