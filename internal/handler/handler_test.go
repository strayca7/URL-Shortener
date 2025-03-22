package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	_ "url-shortener/internal/config"
	"url-shortener/internal/pkg/controller"
	"url-shortener/internal/pkg/database"
	"url-shortener/internal/pkg/middleware"
	"url-shortener/internal/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestRedirect(t *testing.T) {
	pwd, _ := os.Getwd()
	
	log.Info().Msg("当前工作目录: "+pwd)

	database.InitMysqlDB()
	defer database.CloseMysqlDB()

	gin.SetMode(gin.TestMode)

	r := gin.Default()
	r.POST("/register", controller.Register)
	r.POST("/login", controller.Login)

	authGroup := r.Group("/auth")
	authGroup.Use(middleware.JwtAuth())
	{
		authGroup.POST("/short", CreateShorterCodeHandler)
		authGroup.POST("/:code", RedirectHandler)
	}

	t.Run("Successful Registration", func(t *testing.T) {
		body := `{"email": "redirection13@example.com", "password": "P@ssw0rd"}`
		req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		fmt.Println(w.Body.String())
		assert.Contains(t, w.Body.String(), "user_id")
		assert.Contains(t, w.Body.String(), "redirection13@example.com")
	})

	t.Run("Login and Redirect URL", func(t *testing.T) {
		loginBody := `{"email": "redirection13@example.com", "password": "P@ssw0rd"}`
		req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		fmt.Println("Login Response:", w.Body.String())
		assert.Equal(t, http.StatusOK, w.Code)

		var loginResponse util.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
		assert.NoError(t, err, "解析登录响应时不应出错")

		shortURLBody := `{"long_url": "https://www.google.com"}`
		authreq, _ := http.NewRequest("POST", "/auth/short", bytes.NewBufferString(shortURLBody))
		authreq.Header.Set("Content-Type", "application/json")
		authreq.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)
		authreq.Header.Set("refresh_token", loginResponse.RefreshToken)

		w = httptest.NewRecorder()
		r.ServeHTTP(w, authreq)

		fmt.Println("Short URL Response:", w.Body.String())
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "short_url")

		type Code struct {
			ShortURL string `json:"short_url"`
		}

		var shortURL Code
		json.Unmarshal(w.Body.Bytes(), &shortURL)
		redirectreq, _ := http.NewRequest("POST", "/auth/"+shortURL.ShortURL, nil)
		redirectreq.Header.Set("Content-Type", "application/json")
		redirectreq.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)
		redirectreq.Header.Set("refresh_token", loginResponse.RefreshToken)
		redirectreq.RemoteAddr = "192.168.1.1:12345"
		
		w = httptest.NewRecorder()
		r.ServeHTTP(w, redirectreq)
		assert.Equal(t, http.StatusFound, w.Code)
	})
}
