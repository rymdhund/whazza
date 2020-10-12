package base

import (
	"time"
)

type Result struct {
	Status    string
	Msg       string
	Timestamp time.Time
}

func GoodResult() Result {
	return Result{
		Status:    "good",
		Msg:       "",
		Timestamp: time.Now(),
	}
}

func FailResult(msg string) Result {
	return Result{
		Status:    "fail",
		Msg:       msg,
		Timestamp: time.Now(),
	}
}

func ExpiredResult() Result {
	return Result{
		Status:    "expired",
		Msg:       "",
		Timestamp: time.Now(),
	}
}

func NoDataResult() Result {
	return Result{
		Status:    "expired",
		Msg:       "",
		Timestamp: time.Now(),
	}
}
