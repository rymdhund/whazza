package serverdb

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rymdhund/whazza/internal/base"
)

func Init() {
	database, _ := sql.Open("sqlite3", "./whazza.db")

	statement, _ := database.Prepare(`
	CREATE TABLE IF NOT EXISTS results (
		id INTEGER PRIMARY KEY,
		check_id INTEGER,
		status TEXT,
		status_msg TEXT,
		timestamp INTEGER
	)
	`)
	statement.Exec()

	statement, _ = database.Prepare(`
	CREATE TABLE IF NOT EXISTS checks (
		id INTEGER PRIMARY KEY,
		check_type TEXT,
		namespace TEXT,
		params_encoded TEXT,
		interval INTEGER
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

func AddResult(res base.CheckResult) error {
	checkId, err := GetOrCreateCheckId(res.Check)

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

func GetCheckStatus(check base.Check) (base.CheckStatus, error) {
	database, _ := sql.Open("sqlite3", "./whazza.db")
	defer database.Close()

	checkId, err := GetOrCreateCheckId(check)

	if err != nil {
		return base.CheckStatus{}, err
	}

	var (
		lastRes, lastGood, lastFail base.CheckResult
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
		return base.CheckStatus{}, err
	default:
		lastRes.Timestamp = time.Unix(timestamp, 0)
		lastRes.Check = check
	}

	// last good
	err = database.QueryRow(
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? AND status = 'good' ORDER BY timestamp DESC LIMIT 1", checkId,
	).Scan(&lastGood.Status, &lastGood.StatusMsg, &timestamp)
	switch {
	case err == sql.ErrNoRows:
		// empty result
	case err != nil:
		return base.CheckStatus{}, err
	default:
		lastGood.Timestamp = time.Unix(timestamp, 0)
		lastGood.Check = check
	}

	// last fail
	err = database.QueryRow(
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? AND status = 'fail' ORDER BY timestamp DESC LIMIT 1", checkId,
	).Scan(&lastFail.Status, &lastFail.StatusMsg, &timestamp)
	switch {
	case err == sql.ErrNoRows:
		// empty result
	case err != nil:
		return base.CheckStatus{}, err
	default:
		lastFail.Timestamp = time.Unix(timestamp, 0)
		lastFail.Check = check
	}

	var result base.CheckResult
	if (lastRes != base.CheckResult{}) {
		if lastRes.Timestamp.Add(time.Duration(check.Interval) * time.Second).Before(time.Now()) {
			result = base.CheckResult{Check: check, Status: "expired", Timestamp: time.Now()}
		} else {
			result = lastRes
		}
	} else {
		result = base.CheckResult{Check: check, Status: "nodata", Timestamp: time.Now()}
	}

	return base.CheckStatus{Check: check, Result: result, LastReceived: lastRes, LastGood: lastGood, LastFail: lastFail}, nil
}
