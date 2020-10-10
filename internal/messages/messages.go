package messages

import (
	"fmt"

	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/chk"
)

// Messages from agent
type CheckResultMsg struct {
	Check  chk.Check
	Result base.Result
}

func NewCheckResultMsg(check chk.Check, result base.Result) CheckResultMsg {
	return CheckResultMsg{
		Check:  check,
		Result: result,
	}
}

func (cr CheckResultMsg) Validate() (bool, string) {
	if cr.Result.Status != "fail" && cr.Result.Status != "good" {
		return false, fmt.Sprintf("Invalid status: %s", cr.Result.Status)
	}
	return true, ""
}
