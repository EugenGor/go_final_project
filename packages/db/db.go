package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
)

func checkFileExists(dbFile string) bool {
	log.Printf("Check file existance %s", dbFile)

	_, err := os.Stat(dbFile)

	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("DB file %s doesn't exist.", dbFile)
			return false
		}
		log.Fatal(err)
		return false
	}
	log.Printf("DB file %s exists.", dbFile)
	return true
}

func dbCreate(dbFilePath string) {
	// запрос создания таблицы scheduler со столбцами id, date, title, comment, repeat и индексом для date
	taskTableCreateQuery := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date    VARCHAR(8) NOT NULL,
		title   VARCHAR(128) NOT NULL,
		comment VARCHAR(250),
		repeat  VARCHAR(128)
	);
	CREATE INDEX IF NOT EXISTS scheduler_date ON scheduler(date);
	`

	db, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// запрос создания таблицы из taskTableCreateQuery
	_, err = db.Exec(taskTableCreateQuery)
	if err != nil {
		log.Fatal(err)
	}
}

// Подключение к базе данных. Если базы данных нет, то создаем при помощи dbCreate
func DbConnection() {
	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")

	//TODO_DBFILE
	pathDb := os.Getenv("TODO_DBFILE")
	if pathDb != "" {
		dbFile = pathDb
	}

	if !checkFileExists(dbFile) {
		dbCreate(dbFile)
	}
}
