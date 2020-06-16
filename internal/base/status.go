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
	timestring := func(t time.Time) string {
		if (t != time.Time{}) {
			return t.Format(time.RFC3339)
		} else {
			return "N/A"
		}
	}
	return fmt.Sprintf("[%s] <%s> %s, last-res: %s, last-good: %s, last-fail: %s",
		st.Check.Namespace,
		st.Result.Status,
		st.Check.Name(),
		timestring(st.LastReceived.Timestamp),
		timestring(st.LastGood.Timestamp),
		timestring(st.LastFail.Timestamp),
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
	if cr.Result.Status != "fail" && cr.Result.Status != "good" {
		return false, fmt.Sprintf("Invalid status: %s", cr.Result.Status)
	}
	return true, ""
}
