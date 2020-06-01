package serverdb

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rymdhund/whazza/internal/base"
)

func Init() {
	database, _ := sql.Open("sqlite3", "./whazza.db")
	defer database.Close()

	statement, _ := database.Prepare(`
	CREATE TABLE IF NOT EXISTS checks (
		id INTEGER PRIMARY KEY,
		check_type TEXT NOT NULL,
		namespace TEXT NOT NULL,
		params_encoded TEXT NOT NULL,
		interval INTEGER NOT NULL
	)
	`)
	statement.Exec()

	statement, _ = database.Prepare(`
	CREATE TABLE IF NOT EXISTS results (
		id INTEGER PRIMARY KEY,
		check_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		status_msg TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		FOREIGN KEY(check_id) REFERENCES checks(id)
	)
	`)
	statement.Exec()
}

func AddCheck(check base.Check) (int64, error) {
	database, _ := sql.Open("sqlite3", "./whazza.db")
	statement, _ := database.Prepare(
		`INSERT INTO checks
		(check_type, namespace, params_encoded, interval)
		VALUES (?, ?, ?, ?)`)
	res, err := statement.Exec(check.CheckType, check.Namespace, check.ParamsEncoded(), check.Interval)
	if err != nil {
		return 0, err
	}

	id, _ := res.LastInsertId()
	return id, nil
}

func GetOrCreateCheckId(check base.Check) (int64, error) {
	database, _ := sql.Open("sqlite3", "./whazza.db")
	defer database.Close()

	var checkId int64
	err := database.QueryRow(
		"SELECT id FROM checks WHERE check_type = ? AND namespace = ? AND params_encoded = ?",
		check.CheckType,
		check.Namespace,
		check.ParamsEncoded(),
	).Scan(&checkId)
	switch {
	case err == sql.ErrNoRows:
		checkId, err = AddCheck(check)
		if err != nil {
			return 0, err
		}
		return checkId, nil
	case err != nil:
		return 0, err
	default:
		return checkId, nil
	}
}

func AddResult(res base.Result, check base.Check) error {
	checkId, err := GetOrCreateCheckId(check)

	if err != nil {
		return err
	}

	database, _ := sql.Open("sqlite3", "./whazza.db")
	defer database.Close()

	statement, _ := database.Prepare(
		`INSERT INTO results
		(check_id, status, status_msg, timestamp)
		VALUES (?, ?, ?, ?)`)
	_, err = statement.Exec(checkId, res.Status, res.StatusMsg, res.Timestamp.Unix())
	if err != nil {
		return err
	}
	return nil
}

func GetCheckOverview(check base.Check) (overview base.CheckOverview, err error) {
	database, _ := sql.Open("sqlite3", "./whazza.db")
	defer database.Close()

	checkId, err := GetOrCreateCheckId(check)

	if err != nil {
		return
	}

	var (
		lastRes, lastGood, lastFail base.Result
		timestamp                   int64
	)

	// last res
	err = database.QueryRow(
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? ORDER BY timestamp DESC LIMIT 1", checkId,
	).Scan(&lastRes.Status, &lastRes.StatusMsg, &timestamp)
	switch {
	case err == sql.ErrNoRows:
		// empty result
	case err != nil:
		return
	default:
		lastRes.Timestamp = time.Unix(timestamp, 0)
	}

	// last good
	err = database.QueryRow(
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? AND status = 'good' ORDER BY timestamp DESC LIMIT 1", checkId,
	).Scan(&lastGood.Status, &lastGood.StatusMsg, &timestamp)
	switch {
	case err == sql.ErrNoRows:
		// empty result
	case err != nil:
		return
	default:
		lastGood.Timestamp = time.Unix(timestamp, 0)
	}

	// last fail
	err = database.QueryRow(
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? AND status = 'fail' ORDER BY timestamp DESC LIMIT 1", checkId,
	).Scan(&lastFail.Status, &lastFail.StatusMsg, &timestamp)
	switch {
	case err == sql.ErrNoRows:
		// empty result
	case err != nil:
		return
	default:
		lastFail.Timestamp = time.Unix(timestamp, 0)
	}

	var result base.Result
	if (lastRes != base.Result{}) {
		if lastRes.Timestamp.Add(time.Duration(check.Interval) * time.Second).Before(time.Now()) {
			result = base.Result{Status: "expired", Timestamp: time.Now()}
		} else {
			result = lastRes
		}
	} else {
		result = base.Result{Status: "nodata", Timestamp: time.Now()}
	}

	return base.CheckOverview{Check: check, Result: result, LastReceived: lastRes, LastGood: lastGood, LastFail: lastFail}, nil
}
