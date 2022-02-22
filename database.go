package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const (
	sqliteFile    string = "file:./data/ids.db?cache=shared&_journal=WAL"
	CreateDBQuery string = `
	CREATE TABLE IDs (
		PostId varchar(20) NOT NULL PRIMARY KEY,
		Platform text NOT NULL
	);
	delete from IDs;
	`
)

type Database struct {
	sqlDB *sql.DB
}

// var sqlDB *sql.DB

func (db *Database) Init() {
	var err error
	// os.Remove(sqliteFile)

	if db.sqlDB, err = sql.Open("sqlite3", sqliteFile); err != nil {
		log.Fatalln(err)
	}
	defer db.sqlDB.Close()

	db.ReturnAllRecords()
}

func (db Database) CreateUserProfileTable(createDBQuery string) {
	var err error
	if _, err = db.sqlDB.Exec(createDBQuery); err != nil {
		log.Fatalf("%q: %s\n", err, createDBQuery)
		return
	}
}

func (db Database) InsertRows(rows [][]string) (err error) {
	var _sqlDB *sql.DB
	if _sqlDB, err = sql.Open("sqlite3", sqliteFile); err != nil {
		log.Fatalln(err)
	}
	defer _sqlDB.Close()

	var tx *sql.Tx
	tx, err = _sqlDB.Begin()
	if err != nil {
		log.Fatalln(err)
	}

	var stmt *sql.Stmt
	stmt, err = tx.Prepare("INSERT INTO IDs(PostId, Platform) VALUES(?, ?)")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()

	for _, row := range rows {
		_, err = stmt.Exec(row[0], row[1])
		if err != nil {
			log.Fatalln(err)
		}
	}
	return
}

func (db Database) ReturnAllRecords() (records [][]string, err error) {
	var rows *sql.Rows
	rows, err = db.sqlDB.Query("SELECT PostId, Platform FROM IDs")
	if err != nil {
		log.Fatalln(err)
	}
	defer rows.Close()

	for rows.Next() {
		var postId, platform string
		err = rows.Scan(&postId, &platform)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(postId, platform)
		records = append(records, []string{postId, platform})
	}
	err = rows.Err()
	if err != nil {
		log.Fatalln(err)
	}
	return
}

func (db Database) Query(postId string) (platform string) {
	var err error
	var _sqlDB *sql.DB
	if _sqlDB, err = sql.Open("sqlite3", sqliteFile); err != nil {
		log.Fatalln(err)
	}
	defer _sqlDB.Close()

	var stmt *sql.Stmt
	stmt, err = _sqlDB.Prepare("SELECT * FROM IDs WHERE PostId = ?")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()

	var id string
	var row *sql.Row = stmt.QueryRow(postId)
	err = row.Scan(&id, &platform)
	if err != nil && err.Error() != "sql: no rows in result set" {
		log.Fatalln(err)
	}
	log.Println(id, platform)
	_ = id
	return
}

func (db Database) DeleteFrom() {
	var err error
	var _sqlDB *sql.DB
	if _sqlDB, err = sql.Open("sqlite3", sqliteFile); err != nil {
		log.Fatalln(err)
	}
	defer _sqlDB.Close()

	_, err = _sqlDB.Exec("delete from foo")
	if err != nil {
		log.Fatalln(err)
	}

	_, err = _sqlDB.Exec("insert into foo(id, name) values(1, 'foo'), (2, 'bar'), (3, 'baz')")
	if err != nil {
		log.Fatalln(err)
	}

	var rows *sql.Rows
	rows, err = _sqlDB.Query("select id, name from foo")
	if err != nil {
		log.Fatalln(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatalln(err)
	}
}
