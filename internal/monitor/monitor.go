package monitor

import (
	"log"
	"time"

	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/persist"
)

func CheckForExpired() error {
	db, err := persist.Open("./whazza.db")
	if err != nil {
		return err
	}
	defer db.Close()

	expChecks, err := db.GetExpiredChecks()
	if err != nil {
		return err
	}
	for _, chk := range expChecks {
		lastStatus, err := db.LastNotification(chk.ID)
		if err != nil {
			return err
		}

		if lastStatus != "expired" {
			err := notify(db, chk, base.Result{Status: "expired", Timestamp: time.Now()})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func HandleResult(check persist.CheckModel, res persist.ResultModel) error {
	db, err := persist.Open("./whazza.db")
	if err != nil {
		return err
	}
	defer db.Close()

	oldStatus, err := db.LastNotification(check.ID)
	if err != nil {
		return err
	}

	// we treat no notificated statuses yet as "good"
	if oldStatus == "" {
		oldStatus = "good"
	}

	if res.Status != oldStatus {
		notify(db, check, res.Result)
	}
	return nil
}

func notify(db *persist.DB, check persist.CheckModel, res base.Result) error {
	log.Printf("Notification [%s] %+v", res.Status, check)
	err := db.AddNotification(check.ID, res.Status)
	return err
}
