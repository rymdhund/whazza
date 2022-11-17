package chk

import (
	"encoding/json"
	"fmt"

	"github.com/rymdhund/whazza/internal/base"
)

type CheckInChecker struct {
	Name string `json:"name"`
}

func (c CheckInChecker) Title() string {
	return fmt.Sprintf("check-in:%s", c.Name)
}

func (c CheckInChecker) Type() string {
	return "check-in"
}

func (c CheckInChecker) Validate() error {
	return nil
}

func (c CheckInChecker) AsJson() []byte {
	b, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return b
}

func (c CheckInChecker) Run(ctx *Context) base.Result {
	return base.GoodResult()
}
