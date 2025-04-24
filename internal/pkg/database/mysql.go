package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var mysqlDB *gorm.DB

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
	OriginalURL string     `gorm:"type:text;not null"`
	ShortCode   string     `gorm:"type:varchar(10);uniqueIndex;not null"` // 短码6-10位
	ExpireAt    time.Time  `gorm:"index"`                                 // 过期时间索引
	AccessCount int        `gorm:"default:0"`
	ClientIPs   []ClientIP `gorm:"foreignKey:ShortURLID"`           // 一对多关系（一个短链接对应多个IP）
	UserID      string     `gorm:"type:varchar(36);index;not null"` // 外键关联
}

// 客户端IP表结构
type ClientIP struct {
	gorm.Model
	IPAddress  string `gorm:"type:varchar(45);not null"` // IPv4/IPv6地址
	ShortURLID uint   `gorm:"index;not null"`            // 外键关联ShortURL
}

// DB 操作
func InitMysqlDB() error {
	var (
		mydbUser     = viper.GetString("mysql.user")
		mydbPassword = viper.GetString("mysql.password")
		mydbHost     = viper.GetString("mysql.host")
		mydbPort     = viper.GetString("mysql.port")
		mydbName     = viper.GetString("mysql.database")
	)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", mydbUser, mydbPassword, mydbHost, mydbPort, mydbName)

	var err error = nil
	mysqlDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Err(err).Msg("MySQL connection error")
		return err
	}

	sqlDB, err := mysqlDB.DB()
	if err != nil {
		log.Err(err).Msg("Failed to get underlying *sql.DB")
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(3 * time.Hour)

	if err := sqlDB.Ping(); err != nil {
		log.Err(err).Msg("Failed to ping MySQL")
		return err
	}
	log.Debug().Msg("Successfully connected to MySQL")

	// TODO: 使用 job 或者 initContainer 来执行 AutoMigerate，避免每次启动都执行
	if err := mysqlDB.AutoMigrate(&User{}, &ShortURL{}, &ClientIP{}); err != nil {
		log.Err(err).Msg("Failed to migrate MySQL")
	}
	log.Debug().Msg("MySQL migration completed")

	return err
}

func CloseMysqlDB() {
	sqlDB, err := mysqlDB.DB()
	if err != nil {
		log.Err(err).Msg("Failed to get underlying *sql.DB")
		return
	}
	sqlDB.Close()
	log.Debug().Msg("MySQL connection closed")
}

// 用户操作

func CreateUser(user User) error {
	if err := mysqlDB.Create(&user).Error; err != nil {
		return err
	}
	return nil
}

// GetUserByEmail 通过邮箱获取用户
func GetUserByEmail(email string) (User, error) {
	var user User
	if err := mysqlDB.Where("email = ?", email).First(&user).Error; err != nil {
		return User{}, err
	}
	return user, nil
}

// GetOriginalURLByShortCode 通过短码获取原始链接
func GetOriginalURLByShortCode(shortCode string) (string, error) {
	var shortURL ShortURL
	if err := mysqlDB.Where("short_code = ?", shortCode).First(&shortURL).Error; err != nil {
		return "", err
	}

	if shortURL.ExpireAt.Before(time.Now()) {
		log.Info().Msg("Short URL has expired")
		return "", errors.New("short URL has expired")
	}

	return shortURL.OriginalURL, nil
}

// GetShortURLByShortCode 通过短码获取短链接信息
//
// 这里返回的是短链接的所有信息，包括原始链接、短码、过期时间等
func GetURLByShortCode(shortCode string) (ShortURL, error) {
	var shortURL ShortURL
	if err := mysqlDB.Where("short_code = ?", shortCode).First(&shortURL).Error; err != nil {
		log.Err(err).Msg("Short URL not found")
		return ShortURL{}, err
	}

	if shortURL.ExpireAt.Before(time.Now()) {
		log.Warn().Msg("Short URL has expired")
		return ShortURL{}, errors.New("short URL has expired")
	}

	return shortURL, nil
}

// CreateURL 保存短链接口
func CreateShortURL(short ShortURL, c *gin.Context) error {
	if err := mysqlDB.Create(&short).Error; err != nil {
		log.Err(err).Msg("Failed to save short URL")
		return err
	}
	if err := mysqlDB.Create(&ClientIP{IPAddress: c.ClientIP(), ShortURLID: short.ID}).Error; err != nil {
		log.Err(err).Msg("Failed to save client IP")
		return err
	}
	return nil
}

// LogAccess 记录访问信息
func LogAccess(shortCode string, clientIP string) error {
	var (
		shortURL ShortURL
		err      error
	)

	// 查询短链记录
	if err = mysqlDB.Where("short_code = ?", shortCode).First(&shortURL).Error; err != nil {
		log.Err(err).Msg("Short URL not found")
		return err
	}

	// 更新访问计数和 IP 列表
	if err = mysqlDB.Model(&ShortURL{}).Where("short_code = ?", shortCode).Updates(map[string]interface{}{
		"access_count": gorm.Expr("access_count + 1"),
	}).Error; err != nil {
		log.Err(err).Msg("Failed to update access count")
		return err
	}

	if err = SaveClientIP(shortURL.ID, clientIP); err != nil {
		log.Err(err).Msg("Failed to append client IP")
		return err
	}

	return nil
}

func SaveClientIP(shortURLID uint, clientIP string) error {
	var err error
	if err = mysqlDB.Create(&ClientIP{IPAddress: clientIP, ShortURLID: shortURLID}).Error; err != nil {
		return err
	}
	return nil
}

// 通过用户ID获取用户所有短链
func GetUserShortURLsByUserID(userID string) ([]ShortURL, error) {
	var shortURLs []ShortURL
	if err := mysqlDB.Where("user_id = ?", userID).Find(&shortURLs).Error; err != nil {
		log.Err(err).Msg("Failed to get short URLs for userID")
		return nil, err
	}
	return shortURLs, nil
}
