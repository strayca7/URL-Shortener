package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(){
	dsn := "host=localhost user=postgres password=your-password dbname=your-dbname port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		  log.Fatalf("Failed to get underlying *sql.DB: %v", err)
	}
	
	sqlDB.SetMaxIdleConns(10)   // 空闲连接池大小
	sqlDB.SetMaxOpenConns(100)  // 最大打开连接数
	sqlDB.SetConnMaxLifetime(3 *time.Hour) // 连接最大存活时间

    fmt.Println("Database connected successfully!")
}

func CloseDB() {
    sqlDB, err := DB.DB()
    if err != nil {
        log.Printf("Failed to get underlying *sql.DB: %v", err)
        return
    }
    sqlDB.Close()
    log.Println("Database connection closed.")
}