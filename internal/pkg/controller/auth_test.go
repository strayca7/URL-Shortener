package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/handler"
	"url-shortener/internal/pkg/database"
	"url-shortener/internal/pkg/middleware"
	"url-shortener/internal/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	database.InitMysqlDB()
	defer database.CloseMysqlDB()

	gin.SetMode(gin.TestMode)

	r := gin.Default()
	r.POST("/register", Register)

	// 测试用例 1: 参数格式错误
	t.Run("Invalid Request Body", func(t *testing.T) {
		body := `{"email": "invalid-email", "password": "short"}`
		req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "parameters format error")
	})

	// 测试用例 2: 邮箱已注册
	t.Run("Email Already Registered", func(t *testing.T) {
		existingUser := database.User{
			UserID:       "test-user-id",
			Email:        "test@example.com",
			PasswordHash: "hashed-password",
		}
		database.MysqlDB.Create(&existingUser)

		body := `{"email": "test@example.com", "password": "P@ssw0rd"}`
		req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "email already registered")
	})

	// 测试案例 3: 测试注册成功
	t.Run("Successful Registration", func(t *testing.T) {
		body := `{"email": "newuser@example.com", "password": "P@ssw0rd"}`
		req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "user_id")
		assert.Contains(t, w.Body.String(), "newuser@example.com")
	})
}


func TestLogin(t *testing.T) {
    database.InitMysqlDB()
    defer database.CloseMysqlDB()

    gin.SetMode(gin.TestMode)

    r := gin.Default()
    r.POST("/register", Register)
    r.POST("/login", Login)

    authGroup := r.Group("/auth")
    authGroup.Use(middleware.JwtAuth())
    {
        authGroup.POST("/short", handler.CreateShorterCodeHandler)
    }

    // 测试注册成功
    t.Run("Successful Registration", func(t *testing.T) {
        body := `{"email": "newuser3@example.com", "password": "P@ssw0rd"}`
        req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()

        r.ServeHTTP(w, req)

        assert.Equal(t, http.StatusCreated, w.Code)
        fmt.Println(w.Body.String())
        assert.Contains(t, w.Body.String(), "user_id")
        assert.Contains(t, w.Body.String(), "newuser3@example.com")
    })

    // 测试登录并发送 Token 和 long_url
    t.Run("Login and Create Short URL", func(t *testing.T) {
        // 登录请求
        loginBody := `{"email": "newuser3@example.com", "password": "P@ssw0rd"}`
        req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()

        r.ServeHTTP(w, req)

        fmt.Println("Login Response:", w.Body.String())
        assert.Equal(t, http.StatusOK, w.Code)

        // 解析登录响应，提取 Token
        var loginResponse util.LoginResponse
        err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
        assert.NoError(t, err, "解析登录响应时不应出错")

        // 准备创建短链的请求
        shortURLBody := `{"long_url": "https://www.google.com"}`
        authreq, _ := http.NewRequest("POST", "/auth/short", bytes.NewBufferString(shortURLBody))
        authreq.Header.Set("Content-Type", "application/json")
        authreq.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)
        authreq.Header.Set("refresh_token", loginResponse.RefreshToken)

        w = httptest.NewRecorder()
        r.ServeHTTP(w, authreq)

        // 验证响应
        fmt.Println("Short URL Response:", w.Body.String())
        assert.Equal(t, http.StatusOK, w.Code)
        assert.Contains(t, w.Body.String(), "short_url")
    })
}