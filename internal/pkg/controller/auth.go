package controller

import (
	"net/http"
	"url-shortener/internal/pkg/database"
	"url-shortener/internal/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

// Login 将 LoginRequest 对象解析为数据库中的用户对象，并生成 JWT。
//
// Login parses the LoginRequest object into a user object in the database,
// and generates a JWT.
//
// 发送 JSON 格式为： / send JSON format as follows:
//
//	{
//		"email":    "test@example.com",
//		"password": "password123"
//	}
//
// 返回 JSON 格式为： / return JSON format as follows:
//
//	{
//		"access_token": accessToken,
//		"refresh_token": refreshToken,
//		"user": {
//			"user_id": user.UserID,
//			"email":   user.Email,
//		}
//	}
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Err(err).Msg("Request body is invalid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body is invalid"})
		return
	}

	var user database.User
	if err := database.MysqlDB.Where("email = ?", req.Email).First(&user).Error; err == gorm.ErrRecordNotFound {
		log.Debug().Err(err).Msg("Email not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email not found"})
		return
	}

	if !util.ComparePassword(user.PasswordHash, req.Password) {
		log.Debug().Msg("Wrong password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}

	accessToken, refreshToken, err := util.GenerateTokens(user.UserID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"user_id": user.UserID,
			"email":   user.Email,
		},
	})
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

// Register 注册用户，生成数据库对象、UUID。在 http body 以 JSON 格式发送邮箱和密码。
//
// Register registers user, generates database object and UUID. In http body, send email and password in JSON format.
//
// 发送 JSON 格式为： / send JSON format as follows:
//
//	{
//		"email":    "test@example.com",
//		"password": "password123"
//	}
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parameters format error"})
		return
	}

	var existingUser database.User
	if err := database.MysqlDB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	userID := uuid.New().String()

	newUser := database.User{
		UserID:       userID,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}
	if err := database.MysqlDB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_id": userID,
		"email":   req.Email,
	})
}
