package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/hubutil"
	"github.com/rymdhund/whazza/internal/persist"
	"github.com/rymdhund/whazza/internal/sectoken"
)

const dbfile = "./whazza.db"

func startServer() {
	err := hubutil.InitCert()
	if err != nil {
		panic(err)
	}

	db, err := persist.Open(dbfile)
	if err != nil {
		panic(err)
	}
	err = db.Init()
	if err != nil {
		panic(err)
	}
	db.Close()

	http.HandleFunc("/", notFoundHandler)
	http.HandleFunc("/agent/ping", basicAuth(pingHandler))
	http.HandleFunc("/agent/result", basicAuth(resultHandler))
	log.Fatal(http.ListenAndServeTLS(":4433", "cert.pem", "key.pem", nil))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("404: %s", r.URL)
	http.Error(w, "404 Not found", http.StatusNotFound)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		log.Print("Got ping")
		fmt.Fprint(w, "pong")
	default:
		log.Print("Got ping with incorrect method")
		http.Error(w, "405 Method Not allowed", http.StatusMethodNotAllowed)
	}
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		log.Print("Got result")
		decoder := json.NewDecoder(r.Body)
		var checkResult base.CheckResultMsg
		err := decoder.Decode(&checkResult)
		if err != nil {
			log.Printf("Error decoding checkresult: %s", err)
			http.Error(w, "400 Bad Request. Invalid data", http.StatusBadRequest)
			return
		}
		fmt.Printf("Got check result: %+v\n", checkResult)
		ok, e := checkResult.Validate()
		if !ok {
			log.Printf("Invalid checkresult: %s", e)
			http.Error(w, "400 Bad Request. Invalid data", http.StatusBadRequest)
			return
		}
		err = saveResult(checkResult)
		if err != nil {
			log.Printf("Error saving checkresult: %s", e)
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
	default:
		log.Print("Got result with incorrect method")
		http.Error(w, "405 Method Not allowed", http.StatusMethodNotAllowed)
	}
}

func saveResult(res base.CheckResultMsg) error {
	db, err := persist.Open(dbfile)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	err = tx.AddResult(res.Result, res.Check)
	if err != nil {
		tx.Rollback()
		return err
	} else {
		tx.Commit()
	}
	return nil
}

func basicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, rq *http.Request) {
		u, p, ok := rq.BasicAuth()
		if !ok || len(strings.TrimSpace(u)) < 1 || len(strings.TrimSpace(p)) < 1 {
			rw.WriteHeader(http.StatusForbidden)
			return
		}

		t := sectoken.SecToken(p)

		auth, err := authAgent(u, t)
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

func authAgent(name string, token sectoken.SecToken) (bool, error) {
	db, err := persist.Open(dbfile)
	if err != nil {
		return false, err
	}
	defer db.Close()
	return db.AuthenticateAgent(name, token)
}
