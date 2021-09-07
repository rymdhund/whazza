package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/chk"
	"github.com/rymdhund/whazza/internal/hubutil"
	"github.com/rymdhund/whazza/internal/messages"
	"github.com/rymdhund/whazza/internal/monitor"
	"github.com/rymdhund/whazza/internal/persist"
	"github.com/rymdhund/whazza/internal/sectoken"
)

type AuthHandlerFunc func(http.ResponseWriter, *http.Request, persist.AgentModel)

func startServer() {
	err := hubutil.InitCert(Config.KeyFile(), Config.CertFile())
	if err != nil {
		panic(err)
	}

	db, err := persist.Open(Config.Database())
	if err != nil {
		panic(err)
	}
	err = db.Init()
	if err != nil {
		panic(err)
	}
	db.Close()

	mon := monitor.New(Config, time.Now())

	go func() {
		for {
			err := mon.CheckForExpired()
			if err != nil {
				log.Printf("Error in CheckForExpired: %s", err)
			}
			time.Sleep(10 * time.Second)
		}
	}()

	http.HandleFunc("/", notFoundHandler)
	http.HandleFunc("/agent/ping", basicAuth(pingHandler))
	http.HandleFunc("/agent/result", basicAuth(mkResultHandler(mon)))

	addr := fmt.Sprintf(":%d", Config.Port)

	log.Printf("listening on %s", addr)

	log.Fatal(http.ListenAndServeTLS(addr, Config.CertFile(), Config.KeyFile(), nil))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("404: %s", r.URL)
	http.Error(w, "404 Not found", http.StatusNotFound)
}

func pingHandler(w http.ResponseWriter, r *http.Request, agent persist.AgentModel) {
	switch r.Method {
	case "GET":
		log.Print("Got ping")
		fmt.Fprint(w, "pong")
	default:
		log.Print("Got ping with incorrect method")
		http.Error(w, "405 Method Not allowed", http.StatusMethodNotAllowed)
	}
}

func mkResultHandler(mon *monitor.Monitor) AuthHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, agent persist.AgentModel) {
		resultHandler(w, r, agent, mon)
	}
}

func resultHandler(w http.ResponseWriter, r *http.Request, agent persist.AgentModel, mon *monitor.Monitor) {
	switch r.Method {
	case "POST":
		decoder := json.NewDecoder(r.Body)
		var checkResult messages.CheckResultMsg
		err := decoder.Decode(&checkResult)
		if err != nil {
			log.Printf("Error decoding checkresult: %s", err)
			http.Error(w, "400 Bad Request. Invalid data", http.StatusBadRequest)
			return
		}
		fmt.Printf("Got check result: %s - %s\n", checkResult.Check.Title(), checkResult.Result.Status)
		ok, e := checkResult.Validate()
		if !ok {
			log.Printf("Invalid checkresult: %s", e)
			http.Error(w, "400 Bad Request. Invalid data", http.StatusBadRequest)
			return
		}

		err = saveResult(agent, checkResult.Check, checkResult.Result, mon)

		if err != nil {
			log.Printf("Error saving checkresult: %s", err)
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
	default:
		log.Print("Got result with incorrect method")
		http.Error(w, "405 Method Not allowed", http.StatusMethodNotAllowed)
	}
}

func saveResult(agent persist.AgentModel, check chk.Check, result base.Result, mon *monitor.Monitor) error {
	db, err := persist.Open(Config.Database())
	if err != nil {
		return fmt.Errorf("Couldn't open db: %w", err)
	}
	defer db.Close()

	// register check if not exists
	checkModel, err := db.RegisterCheck(agent, check)
	if err != nil {
		return fmt.Errorf("Couldn't register check: %w", err)
	}

	res, err := db.AddResult(agent, checkModel, result)
	if err != nil {
		return fmt.Errorf("Couldn't add result: %w", err)
	}

	go func() {
		err := mon.HandleResult(checkModel, res)
		if err != nil {
			log.Printf("Got error from monitor: %s", err)
		}
	}()

	return nil
}

func basicAuth(handler AuthHandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, rq *http.Request) {
		u, p, ok := rq.BasicAuth()
		if !ok || len(strings.TrimSpace(u)) < 1 || len(strings.TrimSpace(p)) < 1 {
			rw.WriteHeader(http.StatusForbidden)
			return
		}

		t := sectoken.SecToken(p)

		agent, found, err := authAgent(u, t)
		if err != nil {
			log.Printf("Got error authenticating client: %s", err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !found {
			log.Printf("Incorrect login for %s", u)
			rw.WriteHeader(http.StatusForbidden)
			return
		}

		handler(rw, rq, agent)
	}
}

func authAgent(name string, token sectoken.SecToken) (persist.AgentModel, bool, error) {
	db, err := persist.Open(Config.Database())
	if err != nil {
		return persist.AgentModel{}, false, err
	}
	defer db.Close()
	return db.AuthenticateAgent(name, token)
}
