package messages

import (
	"encoding/json"
	"testing"

	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/chk"
)

func TestDeserializeXX(t *testing.T) {
	msg := NewCheckResultMsg(
		chk.Check{},
		base.Result{},
	)
	msg.Check.Type = "http-up"
	bs, _ := json.Marshal(msg)
	json.Unmarshal(bs, msg)
	if msg.Check.Type != "http-up" {
		t.Error("Expected http-up")
	}
}
