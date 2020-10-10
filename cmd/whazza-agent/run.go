package main

import (
	"container/heap"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rymdhund/whazza/internal/agent"
	"github.com/rymdhund/whazza/internal/checking"
	"github.com/rymdhund/whazza/internal/messages"
)

// Priority queue implementation
type timedCheck struct {
	check checking.Check
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

	pq := make(PriorityQueue, len(checks))
	for i, c := range checks {
		pq[i] = &timedCheck{time: time.Now(), check: c}
	}
	heap.Init(&pq)

	if len(pq) < 1 {
		fmt.Printf("No checks to run")
		os.Exit(0)
	}

	for {
		next := pq[0]

		// Sleep until next check is due
		time.Sleep(next.time.Sub(time.Now()))

		fmt.Printf("running check %+v\n", next.check)

		go func() {
			res := next.check.Runner.Run()
			checkResult := messages.NewCheckResultMsg(next.check, res)
			err = agent.SendCheckResult(cfg, checkResult)
			if err != nil {
				log.Printf("Error: couldn't send result: %s", err)
				// TODO: Try again
			}
		}()

		next.time = time.Now().Add(time.Duration(next.check.Interval) * time.Second)
		heap.Fix(&pq, 0)
	}
}
