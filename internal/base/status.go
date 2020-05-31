package base

import (
	"fmt"
	"time"
)

type CheckResult struct {
	Check     Check
	Status    string
	StatusMsg string
	Timestamp time.Time
}

type CheckStatus struct {
	Check        Check
	Result       CheckResult
	LastReceived CheckResult
	LastGood     CheckResult
	LastFail     CheckResult
}

func (st *CheckStatus) Show() string {
	return fmt.Sprintf("%s res: %s, last-res: %s, last-good: %s, last-fail: %s",
		st.Check.Name(),
		st.Result.Status,
		st.LastReceived.Timestamp.Format("15:04"),
		st.LastGood.Timestamp.Format("15:04"),
		st.LastFail.Timestamp.Format("15:04"),
	)
}
