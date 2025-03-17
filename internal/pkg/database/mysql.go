package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var MysqlDB *gorm.DB

type ShortURL struct {
	gorm.Model
	UserID      uint      `gorm:"not null;index"`         // 外键（关联用户）
	OriginalURL string    `gorm:"type:text;not null"`    // 原始 URL
	ShortCode   string    `gorm:"type:varchar(10);uniqueIndex;not null"` // 短链码
	ExpireAt    time.Time // 过期时间
	AccessCount int       `gorm:"default:0"`             // 访问计数
}

type User struct {
	gorm.Model              // 内嵌 gorm.Model（包含 ID、CreatedAt 等字段）
	Name         string    `gorm:"type:varchar(100);not null"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:char(60);not null"` // Bcrypt 哈希值
	ShortURLs    []ShortURL `gorm:"foreignKey:UserID"`     // 一对多关联
}
func InitMysqlDB() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}
	var (
		mydbUser	 = os.Getenv("MYSQL_USER")
		mydbPassword = os.Getenv("MYSQL_PASSWORD")
		mydbHost 	 = os.Getenv("MYSQL_HOST")
		mydbPort 	 = os.Getenv("MYSQL_PORT")
		mydbName 	 = os.Getenv("MYSQL_DATABASE")
	)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", mydbUser, mydbPassword, mydbHost, mydbPort, mydbName)
    var open_err error
	MysqlDB, open_err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if open_err != nil {
        log.Fatal(open_err)
    }
    
	sqlDB, err := MysqlDB.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying *sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(3 * time.Hour)

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping MySQL: %v", err)
	}
	log.Println("MySQL connected successfully!")

	err = MysqlDB.AutoMigrate(&User{}, &ShortURL{})
	if err != nil {
		panic("迁移失败: " + err.Error())
	}
}

func CloseMysqlDB() {
	sqlDB, err := MysqlDB.DB()
	if err != nil {
		log.Printf("Failed to get underlying *sql.DB: %v\n", err)
		return
	}
	sqlDB.Close()
	log.Println("MySQL connection closed.")
}

func GetURL(shortCode string) (string, error) {
	var shortURL ShortURL
	err := MysqlDB.Where("short_code = ?", shortCode).First(&shortURL).Error
	if err != nil {
		return "", err
	}
	return shortURL.OriginalURL, nil
}
func SaveURL(shortCode, longURL string) error {
	return MysqlDB.Create(&ShortURL{OriginalURL: longURL, ShortCode: shortCode}).Error
}

func LogAccess(shortCode string, clientIP string) {
	MysqlDB.Model(&ShortURL{}).Where("short_code = ?", shortCode).Update("access_count", gorm.Expr("access_count + 1"))
}