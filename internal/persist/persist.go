package persist

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3" // sqlite
	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/chk"
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
	connString := fmt.Sprintf("file:%s?_busy_timeout=10000&mode=rwc&_journal_mode=WAL", filename)
	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func OpenMemory() (*DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
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
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS agents (
		id INTEGER PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		token_hash TEXT NOT NULL
	)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS checks (
		id INTEGER PRIMARY KEY,
		agent_id INTEGER NOT NULL,
		type TEXT NOT NULL,
		namespace TEXT NOT NULL,
		interval INTEGER NOT NULL,
		checker_json JSON NOT NULL,
		FOREIGN KEY(agent_id) REFERENCES agents(id)
	)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_checks_big ON checks(agent_id, type, namespace, checker_json)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS results (
		id INTEGER PRIMARY KEY,
		check_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		status_msg TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		FOREIGN KEY(check_id) REFERENCES checks(id)
	)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY,
		check_id INTEGER NOT NULL,
		status TEXT NOT NULL
	)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) AddCheck(agent AgentModel, check chk.Check) (CheckModel, error) {
	res, err := db.Exec(
		`INSERT INTO checks
		(agent_id, type, namespace, interval, checker_json)
		VALUES (?, ?, ?, ?, ?)`,
		agent.ID, check.Type, check.Namespace, check.Interval, check.Checker.AsJson())
	if err != nil {
		return CheckModel{}, err
	}

	id, _ := res.LastInsertId()
	return CheckModel{int(id), check, agent}, nil
}

func (db *DB) RegisterCheck(agent AgentModel, check chk.Check) (CheckModel, error) {
	var (
		checkID  int64
		interval int
	)
	err := db.QueryRow(
		`SELECT id, interval FROM checks WHERE
		  agent_id = ? AND
		  type = ? AND
		  namespace = ? AND
		  checker_json = ?`,
		agent.ID,
		check.Type,
		check.Namespace,
		check.Checker.AsJson(),
	).Scan(&checkID, &interval)
	switch {
	case err == sql.ErrNoRows:
		checkModel, err := db.AddCheck(agent, check)
		if err != nil {
			return CheckModel{}, err
		}
		return checkModel, nil
	case err != nil:
		return CheckModel{}, err
	default:
		// Update interval if changed
		if interval != check.Interval {
			_, err := db.Exec("UPDATE checks SET interval = ? WHERE id = ?", check.Interval, checkID)
			if err != nil {
				return CheckModel{}, err
			}
		}

		return CheckModel{int(checkID), check, agent}, nil
	}
}

func (db *DB) AddResult(agent AgentModel, check CheckModel, res base.Result) (ResultModel, error) {
	r, err := db.Exec(
		`INSERT INTO results
		(check_id, status, status_msg, timestamp)
		VALUES (?, ?, ?, ?)`,
		check.ID, res.Status, res.Msg, res.Timestamp.Unix())
	if err != nil {
		return ResultModel{}, err
	}

	id, _ := r.LastInsertId()

	return ResultModel{int(id), res, check.ID}, nil
}

func (db *DB) AuthenticateAgent(name string, token sectoken.SecToken) (AgentModel, bool, error) {
	var id int

	err := db.QueryRow("SELECT id FROM agents WHERE name = ? AND token_hash = ?", name, token.Hash()).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return AgentModel{}, false, nil
	case err != nil:
		return AgentModel{}, false, err
	default:
		return AgentModel{id, name}, true, nil
	}
}

func (db *DB) SaveAgent(name string, tokenHash string) error {
	_, err := db.Exec(
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
	rows, err := db.Query(
		`SELECT c.id, c.type, c.namespace, c.interval, c.checker_json, a.id, a.name FROM checks c
		JOIN agents a ON c.agent_id = a.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checks := make([]CheckModel, 0)

	for rows.Next() {
		var c CheckModel
		var typ, namespace string
		var interval int
		var jsonData []byte
		err := rows.Scan(
			&c.ID,
			&typ,
			&namespace,
			&interval,
			&jsonData,
			&c.Agent.ID,
			&c.Agent.Name,
		)
		if err != nil {
			return nil, err
		}
		c.Check, err = chk.New(typ, namespace, interval, jsonData)
		if err != nil {
			return nil, err
		}
		checks = append(checks, c)
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
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? ORDER BY timestamp DESC LIMIT 1", check.ID,
	).Scan(&lastRes.Status, &lastRes.Msg, &timestamp)
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
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? AND status = 'good' ORDER BY timestamp DESC LIMIT 1", check.ID,
	).Scan(&lastGood.Status, &lastGood.Msg, &timestamp)
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
		"SELECT status, status_msg, timestamp FROM results WHERE check_id = ? AND status = 'fail' ORDER BY timestamp DESC LIMIT 1", check.ID,
	).Scan(&lastFail.Status, &lastFail.Msg, &timestamp)
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
		now := time.Now()
		if check.Check.IsExpired(lastRes.Timestamp, now) {
			result = base.ExpiredResult()
		} else {
			result = lastRes
		}
	} else {
		result = base.NoDataResult()
	}

	return CheckOverview{CheckModel: check, Result: result, LastReceived: lastRes, LastGood: lastGood, LastFail: lastFail}, nil
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

func (db *DB) GetCheckById(ID int) (CheckModel, error) {
	var c CheckModel
	var typ, namespace string
	var interval int
	var jsonData []byte
	err := db.QueryRow(
		`SELECT c.id, c.type, c.namespace, c.interval, c.checker_json, a.id, a.name
		FROM checks c
		JOIN agents a ON c.agent_id = a.id
		WHERE c.id = ?`,
		ID,
	).Scan(
		&c.ID,
		&typ,
		&namespace,
		&interval,
		&jsonData,
		&c.Agent.ID,
		&c.Agent.Name,
	)
	if err != nil {
		return c, err
	}

	c.Check, err = chk.New(typ, namespace, interval, jsonData)
	if err != nil {
		return c, err
	}

	return c, nil
}

func (db *DB) GetNewerResults(resultID int) ([]ResultModel, error) {
	rows, err := db.Query(
		"SELECT id, status, status_msg, timestamp FROM results WHERE id > ?", resultID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []ResultModel{}

	for rows.Next() {
		var res ResultModel
		var timestamp int64
		err := rows.Scan(&res.ID, &res.Status, &res.Msg, &timestamp)
		if err != nil {
			return nil, err
		}
		res.Timestamp = time.Unix(timestamp, 0)
		results = append(results, res)
	}
	return results, nil
}

// Returns "" if there are no notified statuses
func (db *DB) LastNotification(checkID int) (string, error) {
	var status string
	err := db.QueryRow(
		"SELECT status FROM notifications WHERE check_id = ? ORDER by id DESC",
		checkID,
	).Scan(&status)
	switch {
	case err == sql.ErrNoRows:
		return "", nil
	case err != nil:
		return "", err
	default:
		return status, nil
	}
}

func (db *DB) AddNotification(checkID int, status string) error {
	_, err := db.Exec(
		`INSERT INTO notifications
		(check_id, status)
		VALUES (?, ?)`,
		checkID, status)
	if err != nil {
		return err
	}
	return nil
}
