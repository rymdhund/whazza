package messages

import (
	"fmt"

	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/checking"
)

// Messages from agent
type CheckResultMsg struct {
	Check  checking.Check
	Result base.Result
}

func NewCheckResultMsg(check checking.Check, result base.Result) CheckResultMsg {
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
