package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var PgDB *gorm.DB

func InitPgDB() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}
	var (
		pgdbHost     = os.Getenv("PGSQL_HOST")
		pgdbUser     = os.Getenv("PGSQL_USER")
		pgdbPassword = os.Getenv("PGSQL_PASSWORD")
		pgdbName     = os.Getenv("PGSQL_DBNAME")
		pgdbPort     = os.Getenv("PGSQL_PORT")
		pgdbSSLMode  = os.Getenv("PGSQL_SSLMODE")
		pgdbTimeZone = os.Getenv("PGSQL_TIMEZONE")
	)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		pgdbHost, pgdbUser, pgdbPassword, pgdbName, pgdbPort, pgdbSSLMode, pgdbTimeZone)
	var err error
	PgDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := PgDB.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying *sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)               // 空闲连接池大小
	sqlDB.SetMaxOpenConns(100)              // 最大打开连接数
	sqlDB.SetConnMaxLifetime(3 * time.Hour) // 连接最大存活时间

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}
	log.Println("PostgreSQL connected successfully!")
}

func ClosePgDB() {
	sqlDB, err := PgDB.DB()
	if err != nil {
		log.Printf("Failed to get underlying *sql.DB: %v\n", err)
		return
	}
	sqlDB.Close()
	log.Println("MySQL connection closed.")
}
