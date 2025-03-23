package database

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/datatypes"
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
	OriginalURL string         `gorm:"type:text;not null"`
	ShortCode   string         `gorm:"type:varchar(10);uniqueIndex;not null"` // 短码建议6-10位
	ExpireAt    time.Time      `gorm:"index"`                                 // 过期时间索引
	AccessCount int            `gorm:"default:0"`
	ClientIPs   datatypes.JSON `gorm:"type:json"`
	UserID      string         `gorm:"type:varchar(36);index;not null"` // 外键关联
}

func InitMysqlDB() {
	var (
		mydbUser     = viper.GetString("mysql.user")
		mydbPassword = viper.GetString("mysql.password")
		mydbHost     = viper.GetString("mysql.host")
		mydbPort     = viper.GetString("mysql.port")
		mydbName     = viper.GetString("mysql.database")
	)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", mydbUser, mydbPassword, mydbHost, mydbPort, mydbName)
	var open_err error
	MysqlDB, open_err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if open_err != nil {
		log.Fatal().Err(open_err).Msg("MySQL connection error")
	}

	sqlDB, err := MysqlDB.DB()
	if err != nil {
		log.Err(err).Msg("Failed to get underlying *sql.DB")
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(3 * time.Hour)

	if err := sqlDB.Ping(); err != nil {
		log.Err(err).Msg("Failed to ping MySQL")
	}
	log.Debug().Msg("Successfully connected to MySQL")

	err = MysqlDB.AutoMigrate(&User{}, &ShortURL{})
	if err != nil {
		log.Err(err).Msg("Failed to migrate MySQL")
	}
}

func CloseMysqlDB() {
	sqlDB, err := MysqlDB.DB()
	if err != nil {
		log.Err(err).Msg("Failed to get underlying *sql.DB")
		return
	}
	sqlDB.Close()
	log.Debug().Msg("MySQL connection closed")
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
	clientIP := c.ClientIP()
	var jsonData []byte
	if clientIP != "" {
		jsonData, _ = json.Marshal([]string{clientIP})
	} else {
		jsonData, _ = json.Marshal([]string{})
	}
	return MysqlDB.Create(&ShortURL{OriginalURL: longURL, ShortCode: shortCode, ClientIPs: datatypes.JSON(jsonData)}).Error
}

// LogAccess 记录访问信息
func LogAccess(shortCode string, clientIP string) error {
	var shortURL ShortURL

	// 查询短链记录
	if err := MysqlDB.Where("short_code = ?", shortCode).First(&shortURL).Error; err != nil {
		return err
	}

	var existingIPs []string
	// 解析 JSON 数据到字符串切片
	if len(shortURL.ClientIPs) > 0 {
		if err := json.Unmarshal(shortURL.ClientIPs, &existingIPs); err != nil {
			log.Err(err).Msg("Failed to unmarshal existingIPs")
			existingIPs = []string{} // 初始化为空数组
		}
	}

	// 检查 IP 是否已存在，避免重复记录
	for _, ip := range existingIPs {
		if ip == clientIP {
			// 如果 IP 已存在，只更新访问计数
			return MysqlDB.Model(&ShortURL{}).Where("short_code = ?", shortCode).Update("access_count", gorm.Expr("access_count + 1")).Error
		}
	}

	// 跳过空的 clientIP
	if clientIP == "" {
		log.Debug().Msg("empty clientIP, skip record")
		return nil
	}

	// 添加新的 IP 地址
	existingIPs = append(existingIPs, clientIP)
	updatedIPs, err := json.Marshal(existingIPs)
	if err != nil {
		log.Err(err).Msg("Failed to marshal updatedIPs")
		return err
	}

	// 更新访问计数和 IP 列表
	return MysqlDB.Model(&ShortURL{}).Where("short_code = ?", shortCode).Updates(map[string]interface{}{
		"access_count": gorm.Expr("access_count + 1"),
		"client_ips":   updatedIPs,
	}).Error
}
