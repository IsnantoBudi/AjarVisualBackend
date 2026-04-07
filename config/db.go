package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"ajarvisual-backend/models"

	gormMySQL "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	host := os.Getenv("TIDB_HOST")
	port := os.Getenv("TIDB_PORT")
	user := os.Getenv("TIDB_USER")
	password := os.Getenv("TIDB_PASSWORD")
	database := os.Getenv("TIDB_DATABASE")

	// Validasi: TiDB Cloud Serverless memerlukan format "prefix.username"
	// Contoh yang benar: "2W44eYdvZkFbiuP.root"
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

	db, err := gorm.Open(gormMySQL.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatal("Failed to connect to TiDB: ", err)
	}

	if err := db.AutoMigrate(&models.Worksheet{}); err != nil {
		log.Fatal("Failed to auto migrate: ", err)
	}

	DB = db
	log.Println("Connected to TiDB Cloud")
}
