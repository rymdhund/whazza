package persist

import (
	"fmt"

	. "github.com/rymdhund/whazza/internal/logging"
)

type workWrapper struct {
	work     func(db *DB) error
	response chan error
}

type DbWorker struct {
	workChan chan workWrapper
}

func NewDbWorker() *DbWorker {
	return &DbWorker{
		workChan: make(chan workWrapper),
	}
}

func (w *DbWorker) AddWork(workFunc func(db *DB) error) chan error {
	wrapper := workWrapper{
		work:     workFunc,
		response: make(chan error, 1),
	}
	w.workChan <- wrapper
	return wrapper.response
}

func (w *DbWorker) Run(filename string) {
	go func() {
		db, err := Open(filename)
		if err != nil {
			ErrorLog.Printf("Db worker couldn't open db: %s", err)
			panic(fmt.Errorf("Db worker couldn't open db: %w", err))
		}
		defer db.Close()
		for {
			work, ok := <-w.workChan
			if !ok {
				return
			}
			err := work.work(db)
			work.response <- err
		}
	}()
}
