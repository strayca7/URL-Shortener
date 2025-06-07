package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"url-shortener/internal/pkg/controller"
	"url-shortener/internal/pkg/database"
	"url-shortener/internal/pkg/middleware"
	"url-shortener/internal/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../")
	if err := viper.ReadInConfig(); err != nil {
		panic("Error reading config file")
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    false,
	}

	var multiWriter zerolog.LevelWriter = zerolog.MultiLevelWriter(consoleWriter)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Logger = zerolog.New(multiWriter).
		With().
		Timestamp().
		Logger()

	log.Debug().Msg("Init logger")
	log.Info().Err(errors.New("test error")).Msg("error")
}

func TestUser(t *testing.T) {
	pwd, _ := os.Getwd()

	log.Debug().Msg("当前工作目录: " + pwd)

	database.InitMysqlDB()
	defer database.CloseMysqlDB()

	gin.SetMode(gin.TestMode)

	r := gin.Default()
	r.POST("/register", controller.Register)
	r.POST("/login", controller.Login)

	authGroup := r.Group("/auth")
	authGroup.Use(middleware.JwtAuth())
	{
		authGroup.POST("/short/new", HandleCreateUserShortURL)
		authGroup.POST("/:code", HandleRedirectUserCode)
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
		authreq, _ := http.NewRequest("POST", "/auth/shorten", bytes.NewBufferString(shortURLBody))
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
		// 解析返回短链
		json.Unmarshal(w.Body.Bytes(), &shortURL)
		redirectreq, _ := http.NewRequest("POST", "/auth/short/"+shortURL.ShortURL, nil)
		redirectreq.Header.Set("Content-Type", "application/json")
		redirectreq.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)
		redirectreq.Header.Set("refresh_token", loginResponse.RefreshToken)
		redirectreq.RemoteAddr = "192.168.1.1:12345"

		w = httptest.NewRecorder()
		r.ServeHTTP(w, redirectreq)
		assert.Equal(t, http.StatusFound, w.Code)
	})
}

func TestPublic(t *testing.T) {
	pwd, _ := os.Getwd()

	log.Debug().Msg("当前工作目录: " + pwd)

	database.InitMysqlDB()
	defer database.CloseMysqlDB()

	gin.SetMode(gin.TestMode)

	r := gin.Default()

	public := r.Group("/public")
	{

		public.POST("/register", controller.Register)
		public.POST("/login", controller.Login)
		public.POST("/short/new", HandleCreatePublicShortURL)
		public.GET("/:code", HandleRedirectPublicCode)
		public.GET("/shortcodes", HandleGetAllPublicShortURLs)
	}

	t.Run("Create public URL", func(t *testing.T) {
		bodies := []string{
			`{"long_url": "https://www.google.com"}`,
			`{"long_url": "https://www.baidu.com"}`,
			`{"long_url": "https://www.github.com"}`,
			`{"long_url": "https://www.stackoverflow.com"}`,
		}
		for _, body := range bodies {
			req, _ := http.NewRequest("POST", "/public/short/new", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			fmt.Println(w.Body.String())
			assert.Contains(t, w.Body.String(), "original_url")
			assert.Contains(t, w.Body.String(), "short_url")
		}
	})

	t.Run("Get all public short URLs", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/public/shortcodes", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		fmt.Println(w.Body.String())
	})

	t.Run("Redirect public short URL", func(t *testing.T) {
		// 创建一个新的公共短链
		body := `{"long_url": "https://www.bilibili.com"}`
		req, _ := http.NewRequest("POST", "/public/short/new", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		fmt.Println(w.Body.String())
		assert.Contains(t, w.Body.String(), "original_url")
		assert.Contains(t, w.Body.String(), "short_url")

		var shortURL struct {
			ShortURL string `json:"short_url"`
		}
		json.Unmarshal(w.Body.Bytes(), &shortURL)
		// 解析返回的短链
		redirectreq, _ := http.NewRequest("GET", "/public/"+shortURL.ShortURL, nil)
		redirectreq.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		r.ServeHTTP(w, redirectreq)

		assert.Equal(t, http.StatusFound, w.Code)
	})
}

func TestDelete(t *testing.T) {
	pwd, _ := os.Getwd()

	log.Debug().Msg("当前工作目录: " + pwd)

	database.InitMysqlDB()
	defer database.CloseMysqlDB()

	gin.SetMode(gin.TestMode)

	r := gin.Default()
	publicGroup := r.Group("/public")
	{
		publicGroup.DELETE("/short/:code", HandleDeletePublicShortURL)
	}

	t.Run("Delete short URL", func(t *testing.T) {
		var shortURL struct {
			Code string `json:"code"`
		}
		shortURL.Code = "abc123"
		fmt.Println(shortURL.Code)
		deleteReq, _ := http.NewRequest("DELETE", "/public/short/"+shortURL.Code, nil)
		deleteReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, deleteReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
