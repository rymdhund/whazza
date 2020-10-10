package agent

import (
	"strings"
	"testing"
)

func TestDeserialize(t *testing.T) {
	input := `{
		"checks": [
			{
				"type": "http-up",
				"interval": 30,
				"host": "example.com",
				"port": 12
			}
		]
	}
	`
	checks, err := ParseChecksConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	if len(checks) != 1 {
		t.Fatal("Incorrect number of checks parsed")
	}
	if checks[0].Type != "http-up" {
		t.Fatalf("Incorrect check type")
	}
}

func TestDeserializeError(t *testing.T) {
	input := `{
		"checks": [
			{
				"type": "http-up",
				"interval": 30
			}
		]
	}
	`
	_, err := ParseChecksConfig(strings.NewReader(input))
	if err == nil || err.Error() != "Empty host in http-up check" {
		t.Fatalf("Wrong error: %s", err)
	}
}

func TestDeserializeIgnoreUnknown(t *testing.T) {
	// No simple way to fail on unknown fields since we parse both Check and Checker from the same json
	// We would really like this to fail
	input := `{
		"checks": [
			{
				"type": "http-up",
				"interval": 30,
				"host": "example.com",
				"unknown": 12
			}
		]
	}
	`
	_, err := ParseChecksConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Wrong error: %s", err)
	}
}

func TestDeserializeError3(t *testing.T) {
	input := `{
		"checks": [
			{
				"type": "http-up",
				"host": "example.com",
				"port": "80"
			}
		]
	}
	`
	_, err := ParseChecksConfig(strings.NewReader(input))
	if err == nil {
		t.Fatalf("Wrong error: %s", err)
	}
}
