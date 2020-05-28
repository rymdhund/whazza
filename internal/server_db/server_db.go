package serverdb

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func Init() {
	database, _ := sql.Open("sqlite3", "./whazza.db")

	statement, _ := database.Prepare("""
	CREATE TABLE IF NOT EXISTS results (
		id INTEGER PRIMARY KEY,
		check_id INTEGER,
		status TEXT,
		status_msg TEXT,
		timestamp INTEGER
	)
	""")
	statement.Exec()

	statement, _ := database.Prepare("""
	CREATE TABLE IF NOT EXISTS checks (
		id INTEGER PRIMARY KEY,
		check_type TEXT,
		context TEXT,
		params_encoded TEXT,
		interval INTEGER
	)
	""")
	statement.Exec()


}

func AddResult(res base.Result) {
	database, _ := sql.Open("sqlite3", "./whazza.db")

	rows, _ := database.Query(
		"SELECT id FROM checks WHERE check_type = ? AND context = ? AND params_encoded = ?",
		res.Check.CheckType,
		res.Check.Context,
	)

	// yadayada
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(testStruct)

	if res.Check.

	statement, _ = database.Prepare("INSERT INTO test (a, b) VALUES (?, ?)")
	statement.Exec("x", "y")

	rows, _ := database.Query("SELECT id, a, b FROM test")
	var id int
	var firstname string
	var lastname string
	for rows.Next() {
		rows.Scan(&id, &firstname, &lastname)
		fmt.Println(strconv.Itoa(id) + ": " + firstname + " " + lastname)
	}
}