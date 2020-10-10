package checking

import (
	"bytes"
	"fmt"
	"time"

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

// IsExpired returns true if the check is expired
// We count the check as expired if the check is delayed by Interval/2 (but at least 10m and at most 4h)
func (c Check) IsExpired(lastResult time.Time, now time.Time) bool {
	limit := time.Duration(c.Interval/2) * time.Second
	if limit.Minutes() < 10 {
		limit = time.Duration(10) * time.Minute
	}
	if limit.Hours() > 4 {
		limit = time.Duration(4) * time.Hour
	}
	limit += time.Duration(c.Interval) * time.Second

	return lastResult.Add(limit).Before(now)
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
