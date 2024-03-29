package main

import (
	"container/heap"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rymdhund/whazza/internal/agent"
	"github.com/rymdhund/whazza/internal/chk"
	. "github.com/rymdhund/whazza/internal/logging"
	"github.com/rymdhund/whazza/internal/messages"
)

// Priority queue implementation
type timedCheck struct {
	check chk.Check
	time  time.Time
}

type PriorityQueue []*timedCheck

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].time.Before(pq[j].time)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*timedCheck)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

func run() {
	DebugLog = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLog = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	cfg := readConf()
	f, err := os.Open(checksConfigFile())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading checks.json file: %s\n", err)
		os.Exit(1)
	}

	checks, err := agent.ParseChecksConfig(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading checks.json file: %s\n", err)
		os.Exit(1)
	}
	f.Close()

	hubConn, err := agent.NewHubConnection(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating connection to hub: %s\n", err)
		os.Exit(1)
	}
	checkContext := chk.NewContext()

	pq := make(PriorityQueue, len(checks))
	for i, c := range checks {
		pq[i] = &timedCheck{time: time.Now(), check: c}
	}
	heap.Init(&pq)

	if len(pq) < 1 {
		ErrorLog.Fatalf("No checks to run")
	}

	for {
		next := pq[0]

		// Sleep until next check is due
		time.Sleep(next.time.Sub(time.Now()))

		DebugLog.Printf("running check %+v\n", next.check)

		go func() {
			res := next.check.Checker.Run(checkContext)
			checkResult := messages.NewCheckResultMsg(next.check, res)
			err = hubConn.SendCheckResult(cfg, checkResult)
			if err != nil {
				ErrorLog.Printf("Couldn't send result: %s\n", err)
				// TODO: Try again
			}
		}()

		next.time = time.Now().Add(time.Duration(next.check.Interval) * time.Second)
		heap.Fix(&pq, 0)
	}
}
