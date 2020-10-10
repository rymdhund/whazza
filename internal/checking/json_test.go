package checking

import (
	"encoding/json"
	"testing"

	"github.com/rymdhund/whazza/internal/checking/checker"
)

func TestUnmarshal(t *testing.T) {
	var check Check

	input := `{ "type": "http-up", "interval": 60, "host": "example.com" }`
	err := json.Unmarshal([]byte(input), &check)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if check.Type != "http-up" {
		t.Fatalf("Wrong check type: %s", check.Type)
	}
	if check.Interval != 60 {
		t.Fatalf("Wrong check type: %s", check.Type)
	}
	checker, ok := check.Checker.(checker.HttpUpChecker)
	if !ok {
		t.Fatal("Expected HttpUp")
	}
	if checker.Host != "example.com" {
		t.Fatalf("Wrong host: %s", checker.Host)
	}
}

func TestMarshal(t *testing.T) {
	check := Check{
		checkBase: checkBase{
			Namespace: "a",
			Type:      "http-up",
			Interval:  10,
		},
		Checker: checker.HttpUpChecker{
			Host: "example.com",
		},
	}

	bytes, err := json.Marshal(check)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}

	if string(bytes) != `{"host":"example.com","interval":10,"namespace":"a","type":"http-up"}` {

		t.Fatalf("Unexpected json: %s", string(bytes))
	}

	var check2 Check

	err = json.Unmarshal(bytes, &check2)
	if err != nil {
		t.Fatalf("Errorx: %s", err)
	}

	if check.Type != "http-up" {
		t.Fatalf("Wrong check type: %s", check.Type)
	}
	if check.Interval != 10 {
		t.Fatalf("Wrong interval: %d", check.Interval)
	}
	checker, ok := check.Checker.(checker.HttpUpChecker)
	if !ok {
		t.Fatal("Expected HttpUp")
	}
	if checker.Host != "example.com" {
		t.Fatalf("Wrong host: %s", checker.Host)
	}
}

func TestUnmarshalError(t *testing.T) {
	var check Check

	// No interval
	input := `{ "type": "http-up", "host": "example.com }`
	err := json.Unmarshal([]byte(input), &check)
	if err == nil {
		t.Fatal("Expected error")
	}

	input = `{ "type": "http-up", "interval": 60 }`
	err = json.Unmarshal([]byte(input), &check)
	if err == nil {
		t.Fatal("Expected error")
	}

	input = `{ "type": "http-up", "interval": 60, "host": "example.com" }`
	err = json.Unmarshal([]byte(input), &check)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}
