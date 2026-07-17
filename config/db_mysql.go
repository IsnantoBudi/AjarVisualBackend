//go:build !wasm
// +build !wasm

package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectDB() {
	host := os.Getenv("TIDB_HOST")
	port := os.Getenv("TIDB_PORT")
	user := os.Getenv("TIDB_USER")
	password := os.Getenv("TIDB_PASSWORD")
	database := os.Getenv("TIDB_DATABASE")

	// Validasi: TiDB Cloud Serverless memerlukan format "prefix.username"
	if !strings.Contains(user, ".") {
		log.Fatalf(
			"[TiDB] TIDB_USER tidak valid: '%s'.\n"+
				"TiDB Cloud Serverless memerlukan format 'prefix.username'.\n"+
				"Contoh: '2W44eYdvZkFbiuP.root'\n"+
				"Periksa environment variable di Railway Dashboard Anda.",
			user,
		)
	}

	// TiDB Cloud DSN with TLS
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=true",
		user, password, host, port, database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to open connection to TiDB: ", err)
	}

	// Serverless Connection Pooling Optimization
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect to TiDB: ", err)
	}

	// Create tables if they do not exist
	createWorksheetsQuery := `
	CREATE TABLE IF NOT EXISTS worksheets (
		id INT AUTO_INCREMENT PRIMARY KEY,
		judul_materi VARCHAR(255) NOT NULL,
		tingkat_kelas INT DEFAULT 1,
		data_soal JSON NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createWorksheetsQuery); err != nil {
		log.Fatal("Failed to create worksheets table: ", err)
	}

	createCachedImagesQuery := `
	CREATE TABLE IF NOT EXISTS cached_images (
		prompt_hash VARCHAR(64) PRIMARY KEY,
		prompt TEXT NOT NULL,
		image_data LONGBLOB NOT NULL,
		content_type VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createCachedImagesQuery); err != nil {
		log.Fatal("Failed to create cached_images table: ", err)
	}

	DB = db
	log.Println("Connected to TiDB Cloud (raw database/sql)")
}
