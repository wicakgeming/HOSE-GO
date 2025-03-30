package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"time"

	"backend/config"
	"backend/controllers"
	"backend/middleware"
)

func main() {

	database.ConnectDatabase()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"}, // Sesuaikan dengan frontend
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        AllowCredentials: true, // Jika pakai cookie/token di header
        MaxAge:           12 * time.Hour, // Cache preflight request selama 12 jam
    }))

	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.GET("/protected", func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
	
		c.JSON(http.StatusOK, gin.H{
			"message": "You are authorized",
			"user":    user, // Kirim data user
		})
	})

	r.Run(":8080")
}
