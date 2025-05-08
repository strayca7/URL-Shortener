package handler

import (
	"net/http"
	"sync"

	"url-shortener/internal/pkg/database"
	"url-shortener/internal/pkg/util"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// CreateUserShortURLHandler is an API for creating short URL.
// Requires Authorization and refresh_token in the HTTP header,
// and JSON in the HTTP body.
//
// Send http request, for example: POST http://localhost:8080/auth/short/new
//
// Send http request and JSON format as follows:
//
//	{
//	    "long_url": "https://www.example.com"
//	}
//
// Return JSON format as follows:
//
//	{
//	    "original_url": "https://www.example.com",
//	    "short_url": "abc123"
//	}
//
// The short URL will expire in 90 days. This is default expiration time.
func CreateUserShortURLHandler(c *gin.Context) {
	service.UserShortCodeCreater(c)
}

// RedirectUserCodeHandler handles redirection from a short URL to the original URL.
// Requires Authorization and refresh_token in the HTTP header.
//
// Send http request, for example: POST http://localhost:8080/auth/abc123
func RedirectUserCodeHandler(c *gin.Context) {
	shortCode := c.Param("code")

	originalURL, err := database.GetOriginalURLByShortCode(shortCode)
	if err != nil {
		if err.Error() == "user short URL has expired" {
			log.Warn().Str("shortCode", shortCode).Msg("Short URL has expired")
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		log.Warn().Str("shortCode", shortCode).Msg("Failed to get original URL for shortCode ")
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	clientIP := c.ClientIP()
	log.Info().Str("IP", clientIP).Msg("User IP")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := database.LogUserAccess(shortCode, clientIP); err != nil {
			log.Warn().Str("shortCode", shortCode).Msg("Failed to log access for shortCode ")
		}
	}()
	wg.Wait()

	log.Info().Str("shortCode", shortCode).Str("original URL", originalURL).Msg("Redirecting shortCode ")
	c.Redirect(http.StatusFound, originalURL)
}

// Public short URL redirection handler.
// This handler is used to redirect public short URLs.
// It does not require any authentication or authorization.
func RedirectPublicCodeHandler(c *gin.Context) {
	shortCode := c.Param("code")

	originalURL, err := database.GetPublicShortURLByShortCode(shortCode)
	if err != nil {
		if err.Error() == "public short URL has expired" {
			log.Warn().Str("shortCode", shortCode).Msg("Public short URL has expired")
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		log.Warn().Str("shortCode", shortCode).Msg("Failed to get original URL for shortCode ")
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := database.LogPublicAccess(shortCode); err != nil {
			log.Warn().Str("shortCode", shortCode).Msg("Failed to log access for shortCode ")
		}
	}()
	wg.Wait()
	log.Info().Str("shortCode", shortCode).Str("original URL", originalURL).Msg("Redirecting shortCode ")
	c.Redirect(http.StatusFound, originalURL)
}

// GetUserShortURLsHandler retrieves all short URLs created by the user.
// Requires Authorization and refresh_token in the HTTP header.
//
// It returns a list of all short URLs that owned by the user in JSON format.
func GetUserShortURLsHandler(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		log.Warn().Msg("user ID not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Warn().Msg("error asserting userID to string")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	shortURLs, err := database.GetUserShortURLsByUserID(userIDStr)
	if err != nil {
		log.Warn().Str("userID", userIDStr).Msg("Failed to get short URLs for userID ")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get short URLs"})
	}

	log.Info().Str("userID", userIDStr).Msg("Get short URLs for userID ")
	if len(shortURLs) == 0 {
		log.Info().Str("userID", userIDStr).Msg("No short URLs found for userID ")
		c.JSON(http.StatusOK, gin.H{"message": "No short URLs found"})
		return
	}
	c.JSON(http.StatusOK, shortURLs)
}

// GetAllPublicShortURLsHandler retrieves all public short URLs.
// It does not require any authentication or authorization.
//
// It returns a list of public short URLs that are available to all users in JSON format.
func GetAllPublicShortURLsHandler(c *gin.Context) {
	publicShortURLs, err := database.GetAllPublicShortURLs()
	if err != nil {
		log.Warn().Msg("Failed to get all public short URLs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get all public short URLs"})
		return
	}

	if len(publicShortURLs) == 0 {
		log.Info().Msg("No public short URLs found")
		c.JSON(http.StatusOK, gin.H{"message": "No public short URLs found"})
		return
	}
	c.JSON(http.StatusOK, publicShortURLs)
}

// CreatePublicShortURLHandler is an API for creating public short URL.
// It does not require any authentication or authorization.
//
// Send http request, for example: POST http://localhost:8080/public/short/new
//
// Send JSON format as follows:
//
//	{
//	    "long_url": "https://www.example.com"
//	}
//
// Return JSON format as follows:
//
//	{
//	    "original_url": "https://www.example.com",
//	    "short_url": "abc123"
//	}
//
// The short URL will expire in 90 days. This is default expiration time.
func CreatePublicShortURLHandler(c *gin.Context) {
	service.PublicShortCodeCreater(c)
}

// DeletePublicShortURLHandler is an API for deleting a public short URL.
// It does not require any authentication or authorization.
// Send http request, for example:
//
// DELETE http://localhost:8080/public/short/abc123
//
// It deletes the public short URL with the given short code.
// It returns a success message in JSON format.
func DeletePublicShortURLHandler(c *gin.Context) {
	shortCode := c.Param("code")

	err := database.DeletePublicShortURLByShortCode(shortCode)
	if err != nil {
		log.Warn().Str("shortCode", shortCode).Msg("Failed to delete public short URL")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete public short URL"})
		return
	}

	log.Info().Str("shortCode", shortCode).Msg("Deleted public short URL")
	c.JSON(http.StatusOK, gin.H{"message": "Public short URL deleted successfully"})
}

// RefreshTokenHandler is an API for refreshing the access token.
// It requires Authorization and refresh_token in the HTTP header.
// Send http request, for example: POST http://localhost:8080/auth/refresh
//
// It returns a new access token and refresh token in JSON format.
//
//	{
//	    "access_token": "new_access_token",
//	    "refresh_token": "new_refresh_token"
//	}
func RefreshTokenHandler(c *gin.Context) {
	util.RefreshToken(c)
}
