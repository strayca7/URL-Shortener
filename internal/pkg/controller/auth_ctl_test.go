package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	reqBudy := `{"email": "test@example.com", "password": "P@ssw0rd"}`
	r := gin.Default()
    r.Use(middleware.JwtAuth())
    r.POST("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
    
    // 未携带 Token 的请求
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/register", nil)
    r.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusUnauthorized, w.Code)
    assert.Contains(t, w.Body.String(), "Unauthorized")
}