//go:build wasm
// +build wasm

package config

import (
	"database/sql"
	"log"

	"github.com/syumai/workers/cloudflare/d1"
)

var DB *sql.DB

func ConnectDB() {
	connector, err := d1.OpenConnector("DB")
	if err != nil {
		log.Fatal("Failed to open D1 connector: ", err)
	}
	db := sql.OpenDB(connector)

	// Create tables if they do not exist (SQLite compatible syntax)
	createWorksheetsQuery := `
	CREATE TABLE IF NOT EXISTS worksheets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		judul_materi TEXT NOT NULL,
		tingkat_kelas INTEGER DEFAULT 1,
		data_soal TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createWorksheetsQuery); err != nil {
		log.Fatal("Failed to create worksheets table in D1: ", err)
	}

	createCachedImagesQuery := `
	CREATE TABLE IF NOT EXISTS cached_images (
		prompt_hash TEXT PRIMARY KEY,
		prompt TEXT NOT NULL,
		image_data BLOB NOT NULL,
		content_type TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createCachedImagesQuery); err != nil {
		log.Fatal("Failed to create cached_images table in D1: ", err)
	}

	DB = db
	log.Println("Connected to Cloudflare D1 Database")
}
