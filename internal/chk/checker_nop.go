package chk

import (
	"encoding/json"
	"fmt"

	"github.com/rymdhund/whazza/internal/base"
)

type NopChecker struct {
	Name string `json:"name"`
}

func (c NopChecker) Title() string {
	return fmt.Sprintf("nop:%s", c.Name)
}

func (c NopChecker) Type() string {
	return "nop"
}

func (c NopChecker) Validate() error {
	return nil
}

func (c NopChecker) AsJson() []byte {
	b, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return b
}

func (c NopChecker) Run(ctx *Context) base.Result {
	return base.GoodResult()
}
