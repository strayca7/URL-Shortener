package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var MysqlDB *gorm.DB

// User表结构
type User struct {
	gorm.Model
	UserID       string     `gorm:"type:varchar(36);uniqueIndex;not null"` // UUID格式
	Email        string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string     `gorm:"type:varchar(255);not null"` // 存储bcrypt哈希
	ShortURLs    []ShortURL `gorm:"foreignKey:UserID;references:UserID;onDelete:CASCADE"`
}

// 短链表结构
type ShortURL struct {
	gorm.Model
	OriginalURL string    `gorm:"type:text;not null"`
	ShortCode   string    `gorm:"type:varchar(10);uniqueIndex;not null"` // 短码建议6-10位
	ExpireAt    time.Time `gorm:"index"`                                 // 过期时间索引
	AccessCount int       `gorm:"default:0"`
	UserID      string    `gorm:"type:varchar(36);index;not null"` // 外键关联
}

func InitMysqlDB() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}
	var (
		mydbUser     = os.Getenv("MYSQL_USER")
		mydbPassword = os.Getenv("MYSQL_PASSWORD")
		mydbHost     = os.Getenv("MYSQL_HOST")
		mydbPort     = os.Getenv("MYSQL_PORT")
		mydbName     = os.Getenv("MYSQL_DATABASE")
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
func SaveURL(shortCode string, longURL string, c *gin.Context) error {
	return MysqlDB.Create(&ShortURL{OriginalURL: longURL, ShortCode: shortCode}).Error
}

func LogAccess(shortCode string, clientIP string) {
	MysqlDB.Model(&ShortURL{}).Where("short_code = ?", shortCode).Update("access_count", gorm.Expr("access_count + 1"))
}
