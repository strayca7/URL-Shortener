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

// Login parses the LoginRequest object into a user object in the database,
// and generates a JWT.
//
// Send JSON format as follows:
//
//	{
//		"email":    "test@example.com",
//		"password": "password123"
//	}
//
// Return JSON format as follows:
//
//	{
//		"access_token": accessToken,
//		"refresh_token": refreshToken,
//		"user": {
//			"user_id": user.UserID,
//			"email":   user.Email,
//		}
//	}
//
// TODO: 认证失败次数，ip 存入数据库，限制登录次数
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Info().Err(err).Msg("Request body is invalid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body is invalid"})
		return
	}

	user, err := database.GetUserByEmail(req.Email)
	if err == gorm.ErrRecordNotFound {
		log.Info().Err(err).Msg("Email not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email not found"})
		return
	}
	if err != nil {
		log.Info().Err(err).Msg("Failed to get user by email")
	}

	if !util.ComparePassword(user.PasswordHash, req.Password) {
		log.Info().Msg("Wrong password")
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

// Register registers user, generates database object and UUID. In http body, send email and password in JSON format.
// Send JSON format as follows:
//
//	{
//		"email":    "test@example.com",
//		"password": "password123"
//	}
//
// Return JSON format as follows:
//
//	{
//		"user_id": user.UserID,
//		"email":   user.Email,
//	}
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parameters format error"})
		return
	}

	_, err := database.GetUserByEmail(req.Email)
	if err == gorm.ErrRecordNotFound {
		log.Info().Msg("This email has not been registered, process to register")
	} else if err == nil {
		log.Info().Err(err).Msg("Email already registered")
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	} else {
		log.Info().Err(err).Msg("Failed to get user by email")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
	if err := database.CreateUser(newUser); err != nil {
		log.Warn().Msg("Failed to create user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		return
	}
	log.Info().Msg("User registered successfully")

	c.JSON(http.StatusCreated, gin.H{
		"user_id": userID,
		"email":   req.Email,
	})
}
