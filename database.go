package main

import (
	"fmt"
	"log"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	sqliteFile string = "file:./data/ids.db"
)

type Database struct {
	gormDB *gorm.DB
}

// var db Database

func init() {
	var err error
	if db.gormDB, err = gorm.Open(sqlite.Open(sqliteFile), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}); err != nil {
		log.Fatalln(err)
	}
}

type SavedItem struct { // for saving into gorm
	Id       string
	Platform string
}

func (db Database) CreateTable() {
	db.gormDB.AutoMigrate(&SavedItem{})
}

func (db Database) InsertRow(item SavedItem) {
	// item := SavedItem{Id: newId, Platform: "HackerNews"}
	var result *gorm.DB = db.gormDB.Create(&item)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			log.Printf("result: %+v\n", result.Error.Error())
		} else {
			log.Fatalln(result.Error.Error())
		}
	}
}

func (db Database) InsertRows(items []SavedItem) {
	// item := SavedItem{Id: newId, Platform: "HackerNews"}
	var result *gorm.DB = db.gormDB.Create(&items)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			log.Printf("result: %+v\n", result.Error.Error())
		} else {
			log.Panicln(result.Error.Error())
		}
	}
}

func (db Database) QueryRow(id string) (item SavedItem) {
	item = SavedItem{}
	var result *gorm.DB = db.gormDB.First(&item, "Id = ?", id)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			// log.Printf("result: %+v\n", result.Error.Error())
		} else {
			log.Fatalln(result.Error.Error())
		}
	}
	// if result.RowsAffected != 1 {
	// 	log.Printf("result: %+v\n", result)
	// }
	// log.Println(result.RowsAffected)
	// _ = result
	// log.Printf("user: %+v\n", item)
	return
}

func (db Database) ReturnAllRecords() (savedItems []SavedItem) {
	savedItems = []SavedItem{}
	var result *gorm.DB = db.gormDB.Find(&savedItems)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			log.Printf("result: %+v\n", result.Error.Error())
		} else {
			log.Fatalln(result.Error.Error())
		}
	}
	_ = result
	// log.Println(result.RowsAffected)
	// for _, item := range savedItems {
	// 	log.Printf("user: %+v\n", item)
	// }
	return
}

func (db Database) UpdateXkcd() (item SavedItem) {
	item = SavedItem{}
	// log.Println(&item)
	var result *gorm.DB = db.gormDB.First(&item, "Platform = ?", "xkcd")
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			log.Printf("result: %+v\n", result.Error.Error())
		} else {
			log.Fatalln(result.Error.Error())
		}
	}
	if item == (SavedItem{}) { // if there's no record in the db
		item = SavedItem{Id: "10", Platform: "xkcd"}
		db.InsertRow(item) // create a new record starting from 10
	} else {
		var id int
		id, _ = strconv.Atoi(item.Id)
		item.Id = fmt.Sprint(id + 1)
		db.gormDB.Save(&item) // update id
	}
	// log.Println(result.RowsAffected)
	// _ = result
	// log.Printf("user: %+v\n", item)
	// item.Platform = newPlatform
	// db.gormDB.Save(&item)
	return
}

func (db Database) UpdateRow(targetId, newPlatform string) (item SavedItem) {
	item = SavedItem{}
	var result *gorm.DB = db.gormDB.First(&item, "Id = ?", targetId)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			log.Printf("result: %+v\n", result.Error.Error())
		} else {
			log.Fatalln(result.Error.Error())
		}
	}
	log.Println(result.RowsAffected)
	_ = result
	log.Printf("user: %+v\n", item)
	item.Platform = newPlatform
	db.gormDB.Save(&item)
	return
}

// - -

// package main

// import (
// 	"log"

// 	_ "github.com/mattn/go-sqlite3"
// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"
// 	"gorm.io/gorm/logger"
// )

// const (
// 	sqliteFile    string = "file:./data/ids.db?cache=shared&mode=rwc"
// 	CreateDBQuery string = `
// 	CREATE TABLE IDs (
// 		PostId varchar(20) NOT NULL PRIMARY KEY,
// 		Platform text NOT NULL
// 	);
// 	delete from IDs;
// 	`
// )

//
// sqlite
//

// type Database struct {
// 	// sqlDB  *sql.DB
// 	gormDB *gorm.DB
// }

// var sqlDB *sql.DB

// func (db *Database) Init() {
// 	var err error
// 	// os.Remove(sqliteFile)

// 	if db.sqlDB, err = sql.Open("sqlite3", sqliteFile); err != nil {
// 		log.Fatalln(err)
// 	}
// 	defer db.sqlDB.Close()

// 	db.ReturnAllRecords()
// }

// func (db Database) CreateUserProfileTable(createDBQuery string) {
// 	var err error
// 	if _, err = db.sqlDB.Exec(createDBQuery); err != nil {
// 		log.Fatalf("%q: %s\n", err, createDBQuery)
// 		return
// 	}
// }

// func (db Database) InsertRows(rows [][]string) (err error) {
// 	var _sqlDB *sql.DB
// 	if _sqlDB, err = sql.Open("sqlite3", sqliteFile); err != nil {
// 		log.Fatalln(err)
// 	}
// 	defer _sqlDB.Close()

// 	var tx *sql.Tx
// 	tx, err = _sqlDB.Begin()
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	var stmt *sql.Stmt
// 	stmt, err = tx.Prepare("INSERT INTO IDs(PostId, Platform) VALUES(?, ?)")
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	defer stmt.Close()

// 	for _, row := range rows {
// 		_, err = stmt.Exec(row[0], row[1])
// 		if err != nil {
// 			log.Fatalln(err)
// 		}
// 	}
// 	return
// }

// func (db Database) ReturnAllRecords() (records [][]string, err error) {
// 	var rows *sql.Rows
// 	rows, err = db.sqlDB.Query("SELECT PostId, Platform FROM IDs")
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var postId, platform string
// 		err = rows.Scan(&postId, &platform)
// 		if err != nil {
// 			log.Fatalln(err)
// 		}
// 		log.Println(postId, platform)
// 		records = append(records, []string{postId, platform})
// 	}
// 	err = rows.Err()
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	return
// }

// func (db Database) Query(postId string) (platform string) {
// 	var err error
// 	var _sqlDB *sql.DB
// 	if _sqlDB, err = sql.Open("sqlite3", sqliteFile); err != nil {
// 		log.Fatalln(err)
// 	}
// 	defer _sqlDB.Close()

// 	var stmt *sql.Stmt
// 	stmt, err = _sqlDB.Prepare("SELECT * FROM IDs WHERE PostId = ?")
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	defer stmt.Close()

// 	var id string
// 	var row *sql.Row = stmt.QueryRow(postId)
// 	err = row.Scan(&id, &platform)
// 	if err != nil && err.Error() != "sql: no rows in result set" {
// 		log.Fatalln(err)
// 	}
// 	log.Println(id, platform)
// 	_ = id
// 	return
// }

// func (db Database) DeleteFrom() {
// 	var err error
// 	var _sqlDB *sql.DB
// 	if _sqlDB, err = sql.Open("sqlite3", sqliteFile); err != nil {
// 		log.Fatalln(err)
// 	}
// 	defer _sqlDB.Close()

// 	_, err = _sqlDB.Exec("delete from foo")
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	_, err = _sqlDB.Exec("insert into foo(id, name) values(1, 'foo'), (2, 'bar'), (3, 'baz')")
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	var rows *sql.Rows
// 	rows, err = _sqlDB.Query("select id, name from foo")
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var id int
// 		var name string
// 		err = rows.Scan(&id, &name)
// 		if err != nil {
// 			log.Fatalln(err)
// 		}
// 		log.Println(id, name)
// 	}
// 	err = rows.Err()
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// }

//
// GORM
//

// func init() {
// 	var err error
// 	if db.gormDB, err = gorm.Open(sqlite.Open("./data/ids.db"), &gorm.Config{
// 		Logger: logger.Default.LogMode(logger.Silent),
// 	}); err != nil {
// 		log.Fatalln(err)
// 	}
// }

// type SavedItem struct { // for saving into gorm
// 	Id       string
// 	Platform string
// }

// func (db Database) CreateTable() {
// 	db.gormDB.AutoMigrate(&SavedItem{})
// }

// func (db Database) InsertRow(item SavedItem) {
// 	// user := User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}
// 	var result *gorm.DB = db.gormDB.Create(&item)
// 	_ = result
// 	// log.Printf("%+v\n", result)
// }

// func (db Database) QueryRow(id string) (item SavedItem) {
// 	item = SavedItem{}
// 	var result *gorm.DB = db.gormDB.First(&item, "Id = ?", 1000)
// 	if result.Error != nil {
// 		if result.Error.Error() != "record not found" {
// 			log.Printf("result: %+v\n", result.Error.Error())
// 		} else {
// 			log.Fatalln(result.Error.Error())
// 		}
// 	}
// 	log.Println(result.RowsAffected)
// 	_ = result
// 	log.Printf("user: %+v\n", item)
// 	return
// }
