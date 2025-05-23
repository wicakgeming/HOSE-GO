package middleware

import (
	"net/http"
	"backend/models"
	"backend/config"

	"github.com/gin-gonic/gin"
)

// APIKeyMiddleware - Middleware untuk otorisasi perangkat dengan API Key
// APIKeyMiddleware - Middleware untuk memverifikasi API Key
func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("Authorization")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API Key is required"})
			c.Abort()
			return
		}

		var device models.Device
		if err := database.DB.Where("api_key = ?", apiKey).First(&device).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
			c.Abort()
			return
		}

		// Menyimpan device_id dan api_key ke context
		c.Set("device_id", device.ID)
		c.Set("api_key", apiKey)
		c.Next()
	}
}

