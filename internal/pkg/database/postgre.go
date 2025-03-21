package database

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var PgDB *gorm.DB

func InitPgDB() {
	var (
		pgdbHost     = viper.GetString("pgsql.host")
		pgdbUser     = viper.GetString("pgsql.user")
		pgdbPassword = viper.GetString("pgsql.password")
		pgdbName     = viper.GetString("pgsql.database")
		pgdbPort     = viper.GetString("pgsql.port")
		pgdbSSLMode  = viper.GetString("pgsql.sslmode")
		pgdbTimeZone = viper.GetString("pgsql.timezone")
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
