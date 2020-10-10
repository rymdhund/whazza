package chk

import (
	"testing"
	"time"
)

func MkHttpCheck() Check {
	c, err := New("http-up", "ns", 60, []byte(`{"host":"example.com"}`))
	if err != nil {
		panic(err)
	}
	return c
}

func TestIsExpired(t *testing.T) {
	// If interval is 1 minute, expire after 11 minutes
	check := MkHttpCheck()
	now := time.Now()

	last := now.Add(time.Duration(-(59 + 600)) * time.Second)
	if check.IsExpired(last, now) {
		t.Fatal("Expected not expired")
	}

	last = now.Add(time.Duration(-(61 + 600)) * time.Second)
	if !check.IsExpired(last, now) {
		t.Fatal("Expected expired")
	}

	// If interval is one hour, expire after 90 minutes
	check.Interval = 60 * 60
	last = now.Add(time.Duration(-(90*60 - 1)) * time.Second)
	if check.IsExpired(last, now) {
		t.Fatal("Expected not expired")
	}

	check.Interval = 60 * 60
	last = now.Add(time.Duration(-(90*60 + 1)) * time.Second)
	if !check.IsExpired(last, now) {
		t.Fatal("Expected expired")
	}
}
