package handler

import (
	"log"
	"net/http"
	"sync"

	"url-shortener/internal/pkg/database"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

// CreateShorterCodeHandler 短链生成接口，POST /shorten。
//
// CreateShorterCodeHandler API for creating short URL, POST /shorten.
func CreateShorterCodeHandler(c *gin.Context) {
	service.ShortCodeCreater(c)
}


// RedirectHandler 短链重定向处理函数
//
// RedirectHandler handles redirection from a short URL to the original URL,
// logs the client's IP address, and increments the access count.
func RedirectHandler(c *gin.Context) {
    shortCode := c.Param("code")

    originalURL, err := database.GetURL(shortCode)
    if err != nil {
        log.Printf("Failed to get original URL for shortCode %s: %v\n", shortCode, err)
        c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
        return
    }

    clientIP := c.ClientIP()
	log.Println("用户 IP 为", clientIP)

	var wg sync.WaitGroup
	wg.Add(1)
    go func() {
		defer wg.Done()
        if err := database.LogAccess(shortCode, clientIP); err != nil {
            log.Printf("Failed to log access for shortCode %s: %v\n", shortCode, err)
        }
    }()
	wg.Wait()

    log.Printf("Redirecting shortCode %s to original URL: %s\n", shortCode, originalURL)
    c.Redirect(http.StatusFound, originalURL)
}
