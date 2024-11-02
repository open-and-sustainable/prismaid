package zotero

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func connectDB(dbPath string) *sql.DB {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil
	} else {
		return db
	}
}