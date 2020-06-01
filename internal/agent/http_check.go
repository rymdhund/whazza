package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rymdhund/whazza/internal/base"
)

type HttpCheckParams struct {
	Host        string
	Port        int
	StatusCodes []int
}

func DoCheck(chk base.Check) base.Result {
	var status, msg string

	switch chk.CheckType {
	default:
		panic("invalid check type")
	case "http-up":
		status, msg = DoHttpCheck(chk)
	}
	return base.Result{Status: status, StatusMsg: msg, Timestamp: time.Now()}
}

func DoHttpCheck(chk base.Check) (string, string) {
	params := chk.CheckParams.(HttpCheckParams)
	return HttpCheck(params.Host, params.StatusCodes)
}

func HttpCheck(host string, statusCodes []int) (string, string) {
	resp, err := http.Get(fmt.Sprintf("http://%s/", host))

	if err != nil {
		return "fail", err.Error()
	}
	if statusCodes != nil {
		contains := false
		for _, c := range statusCodes {
			if resp.StatusCode == c {
				contains = true
			}
		}
		if !contains {
			return "fail", fmt.Sprintf("Incorrect http status code: %d", resp.StatusCode)
		}
	}

	return "good", ""
}
