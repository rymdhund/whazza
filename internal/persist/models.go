package persist

import (
	"fmt"
	"time"

	"github.com/rymdhund/whazza/internal/base"
)

type CheckModel struct {
	id int
	base.Check
}

type CheckOverview struct {
	Check        CheckModel
	Result       base.Result
	LastReceived base.Result
	LastGood     base.Result
	LastFail     base.Result
}

func (st *CheckOverview) Show() string {
	timestring := func(t time.Time) string {
		if (t != time.Time{}) {
			return t.Format(time.RFC3339)
		} else {
			return "N/A"
		}
	}
	return fmt.Sprintf("[%s] <%s> %s, last-res: %s, last-good: %s, last-fail: %s, interval: %d",
		st.Check.Namespace,
		st.Result.Status,
		st.Check.Name(),
		timestring(st.LastReceived.Timestamp),
		timestring(st.LastGood.Timestamp),
		timestring(st.LastFail.Timestamp),
		st.Check.Interval,
	)
}
