package checking

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rymdhund/whazza/internal/base"
)

type HttpUpCheckMeta struct{}

type HttpCheckParams struct {
	Host        string
	Port        int
	StatusCodes []int
}

func (m HttpUpCheckMeta) DoCheck(chk base.Check) base.Result {
	var status, msg string

	params, err := m.ParseParams(chk)
	if err != nil {
		panic("couldn't parse params")
	}

	https := chk.CheckType == "https-up"

	status, msg = httpCheck(params.(HttpCheckParams), https)
	return base.Result{Status: status, StatusMsg: msg, Timestamp: time.Now()}
}

func httpCheck(params HttpCheckParams, https bool) (string, string) {
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
		if params.Port == 443 {
			url = fmt.Sprintf("https://%s/", params.Host)
		} else {
			url = fmt.Sprintf("https://%s:%d/", params.Host, params.Port)
		}
	} else {
		if params.Port == 80 {
			url = fmt.Sprintf("http://%s/", params.Host)
		} else {
			url = fmt.Sprintf("http://%s:%d/", params.Host, params.Port)
		}
	}
	resp, err := client.Get(url)

	if err != nil {
		return "fail", err.Error()
	}
	if params.StatusCodes != nil {
		contains := false
		for _, c := range params.StatusCodes {
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

func GetJsonInt(i interface{}) (int, error) {
	switch i.(type) {
	case float64:
		return int(i.(float64)), nil
	default:
		return 0, errors.New("Field is not int")
	}
}

func (m HttpUpCheckMeta) ParseParams(chk base.Check) (interface{}, error) {
	var ret HttpCheckParams

	h, ok := chk.CheckParams["host"]
	if !ok {
		return nil, errors.New("No host in http-up check")
	} else {
		switch h.(type) {
		case string:
			ret.Host = h.(string)
		default:
			return nil, errors.New("Host is not string in http-up check")
		}
	}

	p, ok := chk.CheckParams["port"]
	if !ok {
		if chk.CheckType == "https-up" {
			ret.Port = 443
		} else {
			ret.Port = 80
		}
	} else {
		port, err := GetJsonInt(p)
		if err != nil {
			return nil, errors.New("Invalid port in http-up check")
		}
		ret.Port = port
	}

	sc, ok := chk.CheckParams["status_codes"]
	if !ok {
		// let statuscodes be nil as default
	} else {
		switch sc.(type) {
		case []interface{}:
			sc2 := sc.([]interface{})
			codes := make([]int, len(sc2))
			for i, c := range sc2 {
				n, err := GetJsonInt(c)
				if err != nil {
					return nil, errors.New("Invalid StatusCodes in http-up check")
				}
				codes[i] = n
			}
			ret.StatusCodes = codes
		default:
			return nil, errors.New("StatusCodes is not a list in http-up check")
		}
	}

	return ret, nil
}

func (m HttpUpCheckMeta) DefaultNamespace(chk base.Check) string {
	return chk.CheckParams["host"].(string)
}
