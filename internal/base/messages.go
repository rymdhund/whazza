package base

import (
	"fmt"
)

// Messages from agent
type CheckResultMsg struct {
	Check  Check
	Result Result
}

func (cr CheckResultMsg) Validate() (bool, string) {
	if cr.Check.Interval <= 0 {
		return false, "Invalid interval"
	}
	if cr.Check.CheckType == "" {
		return false, "Empty type"
	}
	if cr.Result.Status != "fail" && cr.Result.Status != "good" {
		return false, fmt.Sprintf("Invalid status: %s", cr.Result.Status)
	}
	return true, ""
}
