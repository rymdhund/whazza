package persist

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/sectoken"
)

type DB struct {
	*sql.DB
}

type Tx struct {
	*sql.Tx
}

// Open returns a DB reference for a data source.
func Open(filename string) (*DB, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// Begin starts an returns a new transaction.
func (db *DB) Begin() (*Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx}, nil
}

func (db *DB) Init() error {
	statement, _ := db.DB.Prepare(`
	CREATE TABLE IF NOT EXISTS checks (
		id INTEGER PRIMARY KEY,
		check_type TEXT NOT NULL,
		namespace TEXT NOT NULL,
		params_encoded TEXT NOT NULL,
		interval INTEGER NOT NULL
	)
	`)
	_, err := statement.Exec()
	if err != nil {
		return err
	}

	statement, _ = db.DB.Prepare(`
	CREATE TABLE IF NOT EXISTS results (
		id INTEGER PRIMARY KEY,
		check_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		status_msg TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		FOREIGN KEY(check_id) REFERENCES checks(id)
	)
	`)
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	statement, _ = db.DB.Prepare(`
	CREATE TABLE IF NOT EXISTS agents (
		id INTEGER PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		token_hash TEXT NOT NULL
	)
	`)
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	return nil
}

func (tx *Tx) AddCheck(check base.Check) (int64, error) {
	res, err := tx.Exec(
		`INSERT INTO checks
		(check_type, namespace, params_encoded, interval)
		VALUES (?, ?, ?, ?)`,
		check.CheckType, check.Namespace, check.ParamsEncoded(), check.Interval)
	if err != nil {
		return 0, err
	}

	id, _ := res.LastInsertId()
	return id, nil
}

func (tx *Tx) RegisterCheck(check base.Check) (int64, error) {
	// TODO: this should take agent as well and use as part of the check key
	var (
		checkId  int64
		interval int
	)
	err := tx.QueryRow(
		"SELECT id, interval FROM checks WHERE check_type = ? AND namespace = ? AND params_encoded = ?",
		check.CheckType,
		check.Namespace,
		check.ParamsEncoded(),
	).Scan(&checkId, &interval)
	switch {
	case err == sql.ErrNoRows:
		checkId, err = tx.AddCheck(check)
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
		_, err := tx.Exec("UPDATE checks SET interval = ? WHERE id = ?", check.Interval, checkId)
		if err != nil {
			return checkId, err
		}
	}

	return checkId, nil
}

func (tx *Tx) AddResult(res base.Result, check base.Check) error {
	checkId, err := tx.RegisterCheck(check)

	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO results
		(check_id, status, status_msg, timestamp)
		VALUES (?, ?, ?, ?)`,
		checkId, res.Status, res.StatusMsg, res.Timestamp.Unix())
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) AuthenticateAgent(name string, token sectoken.SecToken) (bool, error) {
	var id int

	err := db.QueryRow("SELECT id FROM agents WHERE name = ? AND token_hash = ?", name, token.Hash()).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

func (tx *Tx) SaveAgent(name string, tokenHash string) error {
	_, err := tx.Exec(
		`INSERT INTO agents
		(name, token_hash)
		VALUES (?, ?)
		ON CONFLICT(name) DO UPDATE SET token_hash = ?`,
		name, tokenHash, tokenHash)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetChecks() ([]CheckModel, error) {
	rows, _ := db.Query("SELECT id, check_type, namespace, params_encoded, interval FROM checks")
	defer rows.Close()

	checks := make([]CheckModel, 0)

	for rows.Next() {
		var cr CheckModel
		var params []byte
		err := rows.Scan(&cr.id, &cr.CheckType, &cr.Namespace, &params, &cr.Interval)
		if err != nil {
			return nil, err
		}
		cr.CheckParams = base.DecodeParams(params)
		checks = append(checks, cr)
	}
	return checks, nil
}

func (db *DB) GetCheckOverview(check CheckModel) (overview CheckOverview, err error) {
	var (
		lastRes, lastGood, lastFail base.Result
		timestamp                   int64
	)

	// last res
	err = db.QueryRow(
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? ORDER BY timestamp DESC LIMIT 1", check.id,
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
	err = db.QueryRow(
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? AND status = 'good' ORDER BY timestamp DESC LIMIT 1", check.id,
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
	err = db.QueryRow(
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? AND status = 'fail' ORDER BY timestamp DESC LIMIT 1", check.id,
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

	return CheckOverview{Check: check, Result: result, LastReceived: lastRes, LastGood: lastGood, LastFail: lastFail}, nil
}

func (db *DB) GetCheckOverviews() ([]CheckOverview, error) {
	checks, err := db.GetChecks()
	if err != nil {
		return nil, err
	}
	overviews := make([]CheckOverview, 0)
	for _, cr := range checks {
		o, err := db.GetCheckOverview(cr)
		if err != nil {
			return nil, err
		}
		overviews = append(overviews, o)
	}
	return overviews, nil
}
