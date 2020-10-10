package checking

import (
	"encoding/json"
	"testing"
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
	runner, ok := check.Runner.(HttpUpCheck)
	if !ok {
		t.Fatal("Expected HttpUp")
	}
	if runner.Host != "example.com" {
		t.Fatalf("Wrong host: %s", runner.Host)
	}
}

func TestMarshal(t *testing.T) {
	check := Check{
		checkBase: checkBase{
			Namespace: "a",
			Type:      "http-up",
			Interval:  10,
		},
		Runner: HttpUpCheck{
			Host: "example.com",
		},
	}

	bytes, err := json.Marshal(check)
	if err != nil {
		t.Fatalf("Error: %s", err)
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
	runner, ok := check.Runner.(HttpUpCheck)
	if !ok {
		t.Fatal("Expected HttpUp")
	}
	if runner.Host != "example.com" {
		t.Fatalf("Wrong host: %s", runner.Host)
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
