package database

import (
	"time"

	"gorm.io/gorm"
)

// 重构

// User table
type User struct {
	gorm.Model
	UserID       string         `gorm:"type:varchar(36);uniqueIndex;not null"` // UUID格式
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string         `gorm:"type:varchar(255);not null"` // 存储bcrypt哈希
	ShortURLs    []UserShortURL `gorm:"foreignKey:UserID;references:UserID;onDelete:CASCADE"`
}

// User Short URL table
type UserShortURL struct {
	gorm.Model
	OriginalURL string     `gorm:"type:text;not null"`
	ShortCode   string     `gorm:"type:varchar(10);uniqueIndex;not null"` // 短码6-10位
	ExpireAt    time.Time  `gorm:"index"`                                 // 过期时间索引
	AccessCount int        `gorm:"default:0"`
	ClientIPs   []ClientIP `gorm:"foreignKey:ShortURLID"`           // 一对多关系（一个短链接对应多个IP）
	UserID      string     `gorm:"type:varchar(36);index;not null"` // 外键关联
}

// Client IP table
type ClientIP struct {
	gorm.Model
	IPAddress  string `gorm:"type:varchar(45);not null"` // IPv4/IPv6地址
	ShortURLID uint   `gorm:"index;not null"`            // 外键关联UserShortURL
}

// Public Short URL table
type PublicShortURL struct {
	gorm.Model
	ShortCode   string    `gorm:"size:10;uniqueIndex;not null"` // 短链码
	OriginalURL string    `gorm:"type:text;not null"`           // 原始URL
	ExpiresAt   time.Time // 过期时间
	AccessCount uint      `gorm:"default:0"` // 访问计数
}
