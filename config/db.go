package config

import (
"fmt"
"log"
"os"

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

// TiDB Cloud DSN with SSL skip verify via tls param
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
