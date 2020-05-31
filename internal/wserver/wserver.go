package wserver

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rymdhund/whazza/internal/agent"
	"github.com/rymdhund/whazza/internal/base"
	serverdb "github.com/rymdhund/whazza/internal/server_db"
)

func StartServer() {
	serverdb.Init()

	c := base.Check{CheckType: "http-up", Namespace: "net:google.com", CheckParams: agent.HttpCheckParams{"google.com", 80, nil}, Interval: 900}

	//s := base.CheckResult{Check: c, Status: "good", StatusMsg: "", Timestamp: time.Now()}

	//handleMessage(s)
	//handleMessage(base.CheckResult{Check: c, Status: "fail", StatusMsg: "", Timestamp: time.Now()})

	res := agent.DoCheck(c)
	handleMessage(res)

	status, err := serverdb.GetCheckStatus(c)
	if err != nil {
		panic(err)
	}
	fmt.Printf("status: %s\n", status.Show())

	http.HandleFunc("/", webHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleMessage(msg interface{}) {
	fmt.Printf("Got message: %+v\n", msg)

	switch msg := msg.(type) {
	default:
		panic(fmt.Sprintf("unexpected type %T", msg))
	case base.CheckResult:
		err := serverdb.AddResult(msg)
		if err != nil {
			panic(err)
		}
	}
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}
