package base

import (
	"fmt"
	"time"
)

type Result struct {
	Status    string
	StatusMsg string
	Timestamp time.Time
}

type CheckOverview struct {
	Check        Check
	Result       Result
	LastReceived Result
	LastGood     Result
	LastFail     Result
}

func (st *CheckOverview) Show() string {
	return fmt.Sprintf("%s res: %s, last-res: %s, last-good: %s, last-fail: %s",
		st.Check.Name(),
		st.Result.Status,
		st.LastReceived.Timestamp.Format("15:04"),
		st.LastGood.Timestamp.Format("15:04"),
		st.LastFail.Timestamp.Format("15:04"),
	)
}

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
	if cr.Result.Status == "fail" && cr.Result.Status != "good" {
		return false, "Invalid status"
	}
	return true, ""
}
