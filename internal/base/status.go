package base

import (
	"time"
)

type Result struct {
	Status    string
	StatusMsg string
	Timestamp time.Time
}
