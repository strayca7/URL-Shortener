// Package database provides functions to interact with the MySQL, Postgres(not added) database.
// It includes functions to create, read, update, and delete records in the database.
// It also includes functions to initialize and close the database connection.
//
// You can not use this package as a public API to create, read, update, or delete records.
// You should use the handler package instead.
//
// Business log in this package is Debug level. Warning level log will be used in the upper layer.
package database

import (
	"errors"
	"fmt"
	"time"
	"url-shortener/config"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var mysqlDB *gorm.DB

func (u UserShortURL) GetOriginalURL() string {
	return u.OriginalURL
}
func (u UserShortURL) GetShortCode() string {
	return u.ShortCode
}
func (u UserShortURL) GetExpireAt() time.Time {
	return u.ExpireAt
}
func (p PublicShortURL) GetOriginalURL() string {
	return p.OriginalURL
}
func (p PublicShortURL) GetShortCode() string {
	return p.ShortCode
}
func (p PublicShortURL) GetExpireAt() time.Time {
	return p.ExpiresAt
}

const (
	retries         = 25                     // 最大重试次数
	maxRetryDelay   = 100 * time.Millisecond // 最大重试延迟
	maxIdleConns    = 10                     // 最大空闲连接数
	maxOpenConns    = 100                    // 最大打开连接数
	connMaxLifetime = 3 * time.Hour          // 连接最大存活时间
)

type ShowCoder interface {
	GetOriginalURL() string
	GetShortCode() string
	GetExpireAt() time.Time
}

// ###### DB Oprations ######

// InitMysqlDB initializes the MySQL database connection.
//
// Set the maximum idle connections, maximum open connections,
// and connection maximum lifetime.
// It also performs database migrations for the User, UserShortURL, and ClientIP tables.
func InitMysqlDB() {
	log.Info().Msg("** Start init mysql **")

	var (
		mydbUser     = viper.GetString("mysql.user")
		mydbPassword = viper.GetString("mysql.password")
		mydbHost     = viper.GetString("mysql.host")
		mydbPort     = viper.GetString("mysql.port")
		mydbName     = viper.GetString("mysql.database")
	)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", mydbUser, mydbPassword, mydbHost, mydbPort, mydbName)

	var err error
	// try to connect to MySQL, max retries 25 times
	for range retries {
		mysqlDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err == nil {
			break
		}
		time.Sleep(maxRetryDelay)
	}
	if err != nil {
		log.Fatal().Msg("Failed to connect to MySQL, has been retried 100 times.")
	}

	sqlDB, err := mysqlDB.DB()
	if err != nil {
		log.Fatal().Msg("Failed to get underlying *sql.DB.")
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		log.Fatal().Msg("Failed to ping MySQL.")
	}

	log.Info().Msg("Successfully connected to MySQL.")

	// check whether the private tables exist in the database
	if !mysqlDB.Migrator().HasTable(&User{}) || !mysqlDB.Migrator().HasTable(&UserShortURL{}) || !mysqlDB.Migrator().HasTable(&ClientIP{}) {
		log.Info().Msg("Private tables do not exist, starting migration.")
		if err := mysqlDB.AutoMigrate(&User{}, &UserShortURL{}, &ClientIP{}); err != nil {
			log.Err(err).Msg("Failed to migrate MySQL")
		}
	} else {
		log.Info().Msg("Private tables already exist, skipping migration.")
	}

	// check whether the public table exists in the database
	if !mysqlDB.Migrator().HasTable(&PublicShortURL{}) {
		log.Info().Msg("Public public tables do not exist, starting migration.")
		if err := mysqlDB.AutoMigrate(&PublicShortURL{}); err != nil {
			log.Err(err).Msg("Failed to migrate PublicShortURL.")
		}
	} else {
		log.Info().Msg("Public tables already exist, skipping migration.")
	}

	if config.TestMode {
		log.Debug().Msg("Test mode enabled, check tables difference and force migration.")
		if err := mysqlDB.AutoMigrate(&User{}, &UserShortURL{}, &ClientIP{}); err != nil {
			log.Err(err).Msg("Failed to migrate MySQL.")
		}
		if err := mysqlDB.AutoMigrate(&PublicShortURL{}); err != nil {
			log.Err(err).Msg("Failed to migrate PublicShortURL.")
		}
	}

	log.Info().Msg("MySQL migration completed.")

	log.Info().Msg("** Init mysql finished! **")
}

// CloseMysqlDB closes the MySQL database connection.
func CloseMysqlDB() error {
	sqlDB, err := mysqlDB.DB()
	if err != nil {
		log.Err(err).Msg("Failed to get underlying *sql.DB.")
		return err
	}

	if err := sqlDB.Close(); err != nil {
		log.Err(err).Msg("Failed to close MySQL connection.")
		return err
	}
	log.Info().Msg("MySQL connection closed.")
	return nil
}

// ######## User Operations ######

// CreateUser creates a new user in the database.
//
// This function does not judge whether the User already exists.
// You can not use this function as a public API to create a User.
// You should use the Register function instead.
func CreateUser(user User) error {
	if err := mysqlDB.Create(&user).Error; err != nil {
		return err
	}
	return nil
}

// GetUserByEmail retrieves a user by email from the database.
func GetUserByEmail(email string) (User, error) {
	var user User
	if err := mysqlDB.Where("email = ?", email).First(&user).Error; err != nil {
		return User{}, err
	}
	return user, nil
}

// GetOriginalURLByShortCode retrieves the User original URL by short code.
func GetOriginalURLByShortCode(shortCode string) (string, error) {
	var shortURL UserShortURL
	if err := mysqlDB.Where("short_code = ?", shortCode).First(&shortURL).Error; err != nil {
		return "", err
	}

	if shortURL.ExpireAt.Before(time.Now()) {
		log.Debug().Msg("Short URL has expired.")
		return "", errors.New("user short URL has expired")
	}

	return shortURL.OriginalURL, nil
}

// GetURLByShortCode retrieves the User short URL by short code.
//
// If the short code expires, return an error "short URL has expired".
func GetUserShortURLByCode(shortCode string) (UserShortURL, error) {
	var shortURL UserShortURL
	if err := mysqlDB.Where("short_code = ?", shortCode).First(&shortURL).Error; err != nil {
		log.Debug().Msg("User short URL not found.")
		return UserShortURL{}, err
	}

	if shortURL.ExpireAt.Before(time.Now()) {
		log.Debug().Msg("Short URL has expired.")
		return UserShortURL{}, errors.New("short URL has expired")
	}

	return shortURL, nil
}

// CreateUserShortURL creates a new short URL for the user.
func CreateUserShortURL(short UserShortURL, clientIP string) error {
	if err := mysqlDB.Create(&short).Error; err != nil {
		log.Debug().Msg("Failed to save short URL.")
		return err
	}
	if err := mysqlDB.Create(&ClientIP{IPAddress: clientIP, ShortURLID: short.ID}).Error; err != nil {
		log.Debug().Msg("Failed to save client IP.")
		return err
	}
	return nil
}

func ShowCodes(s ShowCoder) (string, string) {
	if s.GetExpireAt().Before(time.Now()) {
		log.Debug().Msg("Short URL has expired.")
		return "", ""
	}
	return s.GetOriginalURL(), s.GetShortCode()
}

// LogUserAccess increments user access count and updates the client IP table.
//
// It will search for the short code in the database before updating the access count.
func LogUserAccess(shortCode string, clientIP string) error {
	var (
		userShortURL UserShortURL
		err          error
	)

	// 查询短链记录
	if err = mysqlDB.Where("short_code = ?", shortCode).First(&userShortURL).Error; err != nil {
		log.Debug().Msg("User short URL not found.")
		return err
	}

	// 更新访问计数和 IP 列表
	if err = mysqlDB.Model(&UserShortURL{}).Where("short_code = ?", shortCode).Updates(map[string]interface{}{
		"access_count": gorm.Expr("access_count + 1"),
	}).Error; err != nil {
		log.Debug().Msg("Failed to update access count.")
		return err
	}

	if err = SaveClientIP(userShortURL.ID, clientIP); err != nil {
		log.Debug().Msg("Failed to append client IP.")
		return err
	}

	return nil
}

// SaveClientIP saves the client IP address to the database.
//
// You can not use this function as a public API to save the client IP.
func SaveClientIP(shortURLID uint, clientIP string) error {
	var err error
	if err = mysqlDB.Create(&ClientIP{IPAddress: clientIP, ShortURLID: shortURLID}).Error; err != nil {
		return err
	}
	return nil
}

// GetUserShortURLsByUserID retrieves all original URLs and short codes
// for a user by user ID. It returns a map of short codes to original URLs.
func GetUserShortURLsByUserID(userID string) (map[string]string, error) {
	var shortURLs []UserShortURL
	if err := mysqlDB.Where("user_id = ?", userID).Find(&shortURLs).Error; err != nil {
		log.Debug().Msg("Failed to get short URLs for userID.")
		return nil, err
	}
	codes := make(map[string]string)
	for _, shortURL := range shortURLs {
		originalURL, shortCode := ShowCodes(shortURL)
		if originalURL != "" && shortCode != "" {
			codes[shortCode] = originalURL
		}
	}
	return codes, nil
}

// ###### Public Operations ######

// LogPublicAccess logs public access count.
//
// If the short code exists, increment the access count by 1.
func LogPublicAccess(shortcode string) error {
	var (
		publicShortURL PublicShortURL
		err            error
	)

	// 查询短链记录
	if err = mysqlDB.Where("short_code = ?", shortcode).First(&publicShortURL).Error; err != nil {
		log.Debug().Msg("Public short URL not found.")
		return err
	}

	// 更新访问计数
	if err = mysqlDB.Model(&PublicShortURL{}).Where("short_code = ?", shortcode).Updates(map[string]interface{}{
		"access_count": gorm.Expr("access_count + 1"),
	}).Error; err != nil {
		log.Debug().Msg("Failed to update access count.")
		return err
	}

	return nil
}

// CreatePublicShortURL creates a new public short URL.
//
// This function does not judge whether the short code already exists.
// You can not use this function as a public API to create a short URL.
// You should use the PublicShortCodeCreater function instead.
func CreatePublicShortURL(short PublicShortURL) error {
	if err := mysqlDB.Create(&short).Error; err != nil {
		log.Debug().Msg("Failed to save public short URL.")
		return err
	}
	return nil
}

// Get a public short URL by short code.
func GetPublicShortURLByShortCode(shortCode string) (string, error) {
	var publicShortURL PublicShortURL
	if err := mysqlDB.Where("short_code = ?", shortCode).First(&publicShortURL).Error; err != nil {
		log.Debug().Msg("Public short URL not found.")
		return "", err
	}

	if publicShortURL.ExpiresAt.Before(time.Now()) {
		log.Debug().Msg("Public short URL has expired.")
		return "", errors.New("public short URL has expired")
	}

	return publicShortURL.OriginalURL, nil
}

// Get all public short URLs.
func GetAllPublicShortURLs() (map[string]string, error) {
	var publicShortURLs []PublicShortURL
	if err := mysqlDB.Find(&publicShortURLs).Error; err != nil {
		log.Debug().Msg("Failed to get all public short URLs.")
		return nil, err
	}
	codes := make(map[string]string)
	for _, publicShortURL := range publicShortURLs {
		originalURL, shortCode := ShowCodes(publicShortURL)
		if originalURL != "" && shortCode != "" {
			codes[shortCode] = originalURL
		}
	}
	return codes, nil
}

// Delete public short URL by short code.
//
// If the short code exists, gorm will not delete it from the database, then it will perform a soft delete instead and set the deleted_at field to the current time.
func DeletePublicShortURLByShortCode(shortCode string) error {
	var publicShortURL PublicShortURL
	if err := mysqlDB.Where("short_code = ?", shortCode).First(&publicShortURL).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug().Msg("Public short URL not found.")
			return err
		} else {
			log.Debug().Msg("Failed to find public short URL.")
			return err
		}
	}

	if err := mysqlDB.Delete(&publicShortURL).Error; err != nil {
		log.Debug().Msg("Failed to delete public short URL.")
		return err
	}
	return nil
}
