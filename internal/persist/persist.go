package persist

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/sectoken"
)

type checkRow struct {
	id    int
	check base.Check
}

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

	statement, _ = database.Prepare(`
	CREATE TABLE IF NOT EXISTS agents (
		id INTEGER PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		token_hash TEXT NOT NULL
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

func RegisterCheck(check base.Check) (int64, error) {
	// TODO: this should take agent as well and use as part of the check key
	database, _ := sql.Open("sqlite3", "./whazza.db")
	defer database.Close()

	var (
		checkId  int64
		interval int
	)
	err := database.QueryRow(
		"SELECT id, interval FROM checks WHERE check_type = ? AND namespace = ? AND params_encoded = ?",
		check.CheckType,
		check.Namespace,
		check.ParamsEncoded(),
	).Scan(&checkId, &interval)
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
	}

	// Update interval if changed
	if interval != check.Interval {
		statement, _ := database.Prepare("UPDATE checks SET interval = ? WHERE id = ?")

		_, err := statement.Exec(check.Interval, checkId)
		if err != nil {
			return checkId, err
		}
	}

	return checkId, nil
}

func AddResult(res base.Result, check base.Check) error {
	checkId, err := RegisterCheck(check)

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

func getChecks() ([]checkRow, error) {
	database, _ := sql.Open("sqlite3", "./whazza.db")
	defer database.Close()

	rows, _ := database.Query("SELECT id, check_type, namespace, params_encoded, interval FROM checks")
	defer rows.Close()

	checks := make([]checkRow, 0)

	for rows.Next() {
		var cr checkRow
		var params []byte
		err := rows.Scan(&cr.id, &cr.check.CheckType, &cr.check.Namespace, &params, &cr.check.Interval)
		if err != nil {
			return nil, err
		}
		cr.check.CheckParams = base.DecodeParams(params)
		checks = append(checks, cr)
	}
	return checks, nil
}

func GetCheckOverviews() ([]base.CheckOverview, error) {
	checks, err := getChecks()
	if err != nil {
		return nil, err
	}
	overviews := make([]base.CheckOverview, 0)
	for _, cr := range checks {
		o, err := getCheckOverview(cr)
		if err != nil {
			return nil, err
		}
		overviews = append(overviews, o)
	}
	return overviews, nil
}

func getCheckOverview(cr checkRow) (overview base.CheckOverview, err error) {
	database, _ := sql.Open("sqlite3", "./whazza.db")
	defer database.Close()

	var (
		lastRes, lastGood, lastFail base.Result
		timestamp                   int64
	)

	// last res
	err = database.QueryRow(
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? ORDER BY timestamp DESC LIMIT 1", cr.id,
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
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? AND status = 'good' ORDER BY timestamp DESC LIMIT 1", cr.id,
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
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? AND status = 'fail' ORDER BY timestamp DESC LIMIT 1", cr.id,
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
		if lastRes.Timestamp.Add(time.Duration(cr.check.Interval) * time.Second).Before(time.Now()) {
			result = base.Result{Status: "expired", Timestamp: time.Now()}
		} else {
			result = lastRes
		}
	} else {
		result = base.Result{Status: "nodata", Timestamp: time.Now()}
	}

	return base.CheckOverview{Check: cr.check, Result: result, LastReceived: lastRes, LastGood: lastGood, LastFail: lastFail}, nil
}

func AuthenticateAgent(name string, token sectoken.SecToken) (bool, error) {
	database, _ := sql.Open("sqlite3", "./whazza.db")
	defer database.Close()

	var id int

	err := database.QueryRow("SELECT id FROM agents WHERE name = ? AND token_hash = ?", name, token.Hash()).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}

}

func SetAgent(name string, tokenHash string) error {
	database, _ := sql.Open("sqlite3", "./whazza.db")
	defer database.Close()

	statement, _ := database.Prepare(
		`INSERT INTO agents
		(name, token_hash)
		VALUES (?, ?)
		ON CONFLICT(name) DO UPDATE SET token_hash = ?`)
	_, err := statement.Exec(name, tokenHash, tokenHash)
	if err != nil {
		return err
	}
	return nil
}
