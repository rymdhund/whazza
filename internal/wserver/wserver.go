package wserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/rymdhund/whazza/internal/agent"
	"github.com/rymdhund/whazza/internal/base"
	serverdb "github.com/rymdhund/whazza/internal/server_db"
	"github.com/rymdhund/whazza/internal/token"
)

func StartServer() {
	err := generateCertIfNotExists()
	if err != nil {
		panic(err)
	}

	serverdb.Init()

	c := base.Check{CheckType: "http-up", Namespace: "net:google.com", CheckParams: agent.HttpCheckParams{"google.com", 80, nil}, Interval: 900}

	res := agent.DoCheck(c)
	checkResult := base.CheckResultMsg{Check: c, Result: res}
	handleMessage(checkResult)

	overview, err := serverdb.GetCheckOverview(c)
	if err != nil {
		panic(err)
	}
	fmt.Printf("status: %s\n", overview.Show())

	http.HandleFunc("/", notFoundHandler)
	http.HandleFunc("/agent/ping", BasicAuth(pingHandler))
	http.HandleFunc("/agent/result", BasicAuth(resultHandler))
	log.Fatal(http.ListenAndServeTLS(":4433", "cert.pem", "key.pem", nil))
}

func handleMessage(msg interface{}) {
	fmt.Printf("Got message: %+v\n", msg)

	switch msg := msg.(type) {
	default:
		panic(fmt.Sprintf("unexpected type %T", msg))
	case base.CheckResultMsg:
		err := serverdb.AddResult(msg.Result, msg.Check)
		if err != nil {
			panic(err)
		}
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 Not found", http.StatusNotFound)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		log.Print("Got ping")
		fmt.Fprint(w, "pong")
	default:
		http.Error(w, "405 Method Not allowed", http.StatusMethodNotAllowed)
	}
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		decoder := json.NewDecoder(r.Body)
		var checkResult base.CheckResultMsg
		err := decoder.Decode(&checkResult)
		if err != nil {
			http.Error(w, "400 Bad Request. Invalid data", http.StatusBadRequest)
			return
		}
		fmt.Printf("Got check result: %+v\n", checkResult)
		ok, _ := checkResult.Validate()
		if !ok {
			http.Error(w, "400 Bad Request. Invalid data", http.StatusBadRequest)
			return
		}
		err = serverdb.AddResult(checkResult.Result, checkResult.Check)
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "405 Method Not allowed", http.StatusMethodNotAllowed)
	}
}

func ShowFingerprint() {
	err := generateCertIfNotExists()
	if err != nil {
		panic(err)
	}

	fp, err := ReadFingerprint()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Cert fingerprint: %s\n", fp)
}

func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, rq *http.Request) {
		u, p, ok := rq.BasicAuth()
		if !ok || len(strings.TrimSpace(u)) < 1 || len(strings.TrimSpace(p)) < 1 {
			rw.WriteHeader(http.StatusForbidden)
			return
		}

		t := token.Token(p)

		auth, err := serverdb.AuthenticateAgent(u, t)
		if err != nil {
			log.Printf("Got error authenticating client: %s", err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !auth {
			log.Printf("Incorrect login for %s", u)
			rw.WriteHeader(http.StatusForbidden)
			return
		}

		handler(rw, rq)
	}
}

func RegisterAgent(name, tokenHash string) {
	err := serverdb.SetAgent(name, tokenHash)
	if err != nil {
		panic(err)
	}
}
