package checking

import (
	"bytes"
	"fmt"

	"github.com/rymdhund/whazza/internal/base"
)

type Check struct {
	checkBase
	Runner CheckRunner
}

type checkBase struct {
	Type      string `json:"type"`
	Namespace string `json:"namespace"`
	Interval  int    `json:"interval"`
}

type CheckRunner interface {
	Type() string
	Name() string
	Run() base.Result
	Validate() error
	AsJson() []byte
}

func (c Check) Validate() error {
	if c.Interval <= 0 {
		return fmt.Errorf("Invalid interval: %d", c.Interval)
	}
	return c.Runner.Validate()
}

func (c Check) Name() string {
	return c.Runner.Name()
}

func New(checkType, namespace string, interval int, runnerJson []byte) (Check, error) {
	check := Check{
		checkBase{
			Type:      checkType,
			Namespace: namespace,
			Interval:  interval,
		},
		nil,
	}

	runner, err := unmarshalRunner(checkType, runnerJson)
	if err != nil {
		return check, err
	}
	check.Runner = runner
	return check, nil
}

func Equal(c1, c2 Check) bool {
	return c1.checkBase == c2.checkBase && bytes.Equal(c1.Runner.AsJson(), c2.Runner.AsJson())
}
