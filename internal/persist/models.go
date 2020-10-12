package persist

import (
	"fmt"
	"time"

	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/chk"
	"github.com/rymdhund/whazza/internal/utils"
)

type AgentModel struct {
	ID   int
	Name string
}

type CheckModel struct {
	ID    int
	Check chk.Check
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
	now := time.Now()

	extra := ""
	if o.Result.Status == "fail" {
		extra = fmt.Sprintf(" | %s | last good: %s", o.Result.Msg, utils.HumanRelTime(now, o.LastGood.Timestamp, false))
	}

	return fmt.Sprintf("[%s] %s | %s | %s%s",
		o.CheckModel.Check.Namespace,
		o.Result.Status,
		o.CheckModel.Check.Name(),
		utils.HumanRelTime(now, o.LastReceived.Timestamp, false),
		extra,
	)
}
