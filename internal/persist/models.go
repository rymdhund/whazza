package persist

import (
	"fmt"
	"time"

	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/checking"
)

type AgentModel struct {
	ID   int
	Name string
}

type CheckModel struct {
	ID    int
	Check checking.Check
	Agent AgentModel
}

type ResultModel struct {
	ID int
	base.Result
	CheckID int
}

type CheckOverview struct {
	CheckModel   CheckModel
	Result       base.Result
	LastReceived base.Result
	LastGood     base.Result
	LastFail     base.Result
}

func (o *CheckOverview) Show() string {
	timestring := func(t time.Time) string {
		if (t != time.Time{}) {
			return t.Format(time.RFC3339)
		} else {
			return "N/A"
		}
	}
	return fmt.Sprintf("[%s] <%s> %s, last-res: %s, last-good: %s, last-fail: %s, interval: %d",
		o.CheckModel.Check.Namespace,
		o.Result.Status,
		o.CheckModel.Check.Name(),
		timestring(o.LastReceived.Timestamp),
		timestring(o.LastGood.Timestamp),
		timestring(o.LastFail.Timestamp),
		o.CheckModel.Check.Interval,
	)
}
