package controller

import (
	"net/http"
	"url-shortener/internal/pkg/database"
	"url-shortener/internal/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

// Login 将 LoginRequest 对象解析为数据库中的用户对象，并生成 JWT。
// Login parses the LoginRequest object into a user object in the database, and generates a JWT.
// 
// 返回 Json 格式为： / return Json format as follows:
// {
// 	"access_token": accessToken,
// 	"refresh_token": refreshToken,
// 	"user": {
// 		"user_id": user.UserID,
// 		"email":   user.Email,
// 	},
// }
//
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parameters format error"})
		return
	}

	var user database.User
	if err := database.MysqlDB.Where("email = ?", req.Email).First(&user).Error; err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email"})
		return
	}

	if !util.ComparePassword(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}

	accessToken, refreshToken, err := util.GenerateTokens(user.UserID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
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

// Register 注册用户，生成数据库对象、UUID。
//
// Register registers user, generates database object and UUID.
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
