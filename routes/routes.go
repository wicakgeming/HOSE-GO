package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"backend/controllers"
	"backend/middleware"
	"backend/models"
	"backend/config"
)

// SetupRouter mengatur semua route untuk aplikasi
func SetupRouter() *gin.Engine {
	// Inisialisasi database
	database.ConnectDatabase()

	// Auto Migrate Model
	database.DB.AutoMigrate(&models.User{}, &models.Device{}, &models.SensorData{})

	// Membuat instance gin router
	r := gin.Default()

	// Pengaturan CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Sesuaikan dengan frontend
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// =================== Public Routes (Tanpa JWT) ===================
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	// =================== Protected Routes (Memerlukan JWT) ===================
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())

	// Endpoint untuk mengecek otorisasi
	protected.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		c.JSON(http.StatusOK, gin.H{
			"message":  "You are authorized",
			"user_id":  userID,
			"username": username,
		})
	})

	// Device Routes
	protected.POST("/device", controllers.CreateDevice)
	protected.GET("/devices/:user_id", controllers.GetDevicesByUser)
	protected.PUT("/device/:device_id", controllers.UpdateDevice)

	// Sensor Routes
	protected.POST("/sensor", controllers.AddSensorData)
	protected.GET("/sensor/:device_id", controllers.GetSensorData)

	// Device API dengan API Key (untuk ESP32-S3)
	deviceAPI := r.Group("/api/device")
	deviceAPI.Use(middleware.APIKeyMiddleware())
	deviceAPI.POST("/sensor", controllers.AddSensorData)
	deviceAPI.GET("/sensor/:device_id", controllers.GetSensorData)


	return r
}
