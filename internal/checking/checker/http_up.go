package checker

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rymdhund/whazza/internal/base"
)

type HttpUpChecker struct {
	Host        string `json:"host"`
	Port        int    `json:"port,omitempty"`
	StatusCodes []int  `json:"status_codes,omitempty"`
}

type HttpsUpChecker struct {
	HttpUpChecker
}

func (c HttpUpChecker) Name() string {
	if c.PortOrDefault() != 80 {
		return fmt.Sprintf("http:%s:%d", c.Host, c.PortOrDefault())
	}
	return fmt.Sprintf("http:%s", c.Host)
}

func (c HttpUpChecker) Type() string {
	return "http-up"
}

func (c HttpUpChecker) Validate() error {
	if c.Host == "" {
		return errors.New("Empty host in http-up check")
	}
	return nil
}

func (c HttpUpChecker) AsJson() []byte {
	b, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return b
}

func (c HttpUpChecker) PortOrDefault() int {
	if c.Port == 0 {
		return 80
	}
	return c.Port
}

func (c HttpUpChecker) Run() base.Result {
	status, msg := httpCheck(c.Host, c.PortOrDefault(), c.StatusCodes, false)
	return base.Result{Status: status, StatusMsg: msg, Timestamp: time.Now()}
}

///////////////////
// Https methods //
///////////////////

func (c HttpsUpChecker) Type() string {
	return "https-up"
}

func (c HttpsUpChecker) Name() string {
	if c.PortOrDefault() != 443 {
		return fmt.Sprintf("https:%s:%d", c.Host, c.PortOrDefault())
	}
	return fmt.Sprintf("https:%s", c.Host)
}

func (c HttpsUpChecker) PortOrDefault() int {
	if c.Port == 0 {
		return 443
	}
	return c.Port
}

func (c HttpsUpChecker) Run() base.Result {
	status, msg := httpCheck(c.Host, c.PortOrDefault(), c.StatusCodes, true)
	return base.Result{Status: status, StatusMsg: msg, Timestamp: time.Now()}
}

func httpCheck(host string, port int, statusCodes []int, https bool) (string, string) {
	// Dont follow redirects and allow bad certs
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	var url string
	if https {
		if port == 443 {
			url = fmt.Sprintf("https://%s/", host)
		} else {
			url = fmt.Sprintf("https://%s:%d/", host, port)
		}
	} else {
		if port == 80 {
			url = fmt.Sprintf("http://%s/", host)
		} else {
			url = fmt.Sprintf("http://%s:%d/", host, port)
		}
	}
	resp, err := client.Get(url)

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
	} else {
		if resp.StatusCode != http.StatusOK {
			return "fail", fmt.Sprintf("Incorrect http status code: %d", resp.StatusCode)
		}
	}

	return "good", ""
}
