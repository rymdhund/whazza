package wserver

import (
	"fmt"
	"time"
	"net/http"
	"log"

	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/server_db"
	"github.com/rymdhund/whazza/internal/agent"
)

type Server struct {
	History map[string] []base.CheckResult
}

func NewServer() Server {
	return Server{History: make(map[string] []base.CheckResult)}
}

func (srv *Server) GetCheckStatus(check base.Check) base.CheckStatus {
	results := srv.History[check.Key()]

	var lastRes, lastGood, lastFail base.CheckResult
	for _, res := range results {
		lastRes = res
		if res.Status == "good" {
			lastGood = res
		} else {
			lastFail = res
		}
	}
	var result base.CheckResult
	if len(results) > 0 {
		if lastRes.Timestamp.Add(time.Duration(check.Interval) * time.Second).Before(time.Now()) {
			result = base.CheckResult{Check: check, Status: "expired", Timestamp: time.Now()}
		} else {
			result = lastRes
		}
	} else {
		result = base.CheckResult{Check: check, Status: "nodata", Timestamp: time.Now()}
	}

	return base.CheckStatus{ Check: check, Result: result, LastReceived: lastRes, LastGood: lastGood, LastFail: lastFail} 
}

func StartServer() {
	c := base.Check{CheckType: "http-up", Context: "net:internet", CheckParams: agent.HttpCheckParams{"google.com", 80, nil}, Interval: 900}

	s := base.CheckResult{Check: c, Status: "good", StatusMsg: "", Timestamp: time.Now()}

	srv := NewServer()
	srv.handle(s)
	srv.handle(base.CheckResult{Check: c, Status: "fail", StatusMsg: "", Timestamp: time.Now()})
	
	res := agent.DoCheck(c)
	srv.handle(res)

	status := srv.GetCheckStatus(c)
	fmt.Printf("status: %s\n", status.Show())

	fmt.Println("hello")
	serverdb.Init()

	http.HandleFunc("/", webHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (srv *Server) handle(msg interface{}) {
	fmt.Printf("Got message: %+v\n", msg)

	switch msg := msg.(type) {
	default:
		panic(fmt.Sprintf("unexpected type %T", msg))
	case base.CheckResult:
		key := msg.Check.Key()

		hist := make([]base.CheckResult, 0)
		h, prs := srv.History[key]
		if prs {
			hist = h
		}
		srv.History[key] = append(hist, msg)
		fmt.Printf("status %+t\n", msg)
		fmt.Printf("history: %v\n", srv.History[key])
	}
}

func webHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}