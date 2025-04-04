package routes

import (
	"net/http"
	"time"

	database "backend/config"
	"backend/controllers"
	"backend/middleware"
	"backend/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	// Device Routes (User)
	protected.GET("/devices", controllers.GetDevicesByUser)
	protected.PUT("/device/:device_id", controllers.UpdateDevice)

	// Sensor Routes
	protected.POST("/sensor", controllers.AddSensorData)
	protected.GET("/sensor/:device_id", controllers.GetSensorData)

	// Device API dengan API Key (untuk ESP32-S3)
	deviceAPI := r.Group("/api/device")
	deviceAPI.Use(middleware.APIKeyMiddleware())
	deviceAPI.POST("/sensor", controllers.AddSensorData)
	deviceAPI.GET("/sensor/:device_id", controllers.GetSensorData)
	deviceAPI.GET("/status", controllers.GetDeviceStatus) // Endpoint untuk melihat status device

	// Admin Routes (Hanya bisa diakses admin)
	protectedAdmin := r.Group("/admin")
	protectedAdmin.Use(middleware.AuthMiddleware(), middleware.AdminOnly())

	// Routes untuk User Management (Hanya Admin)
	protectedAdmin.POST("/users", controllers.CreateUser)       // Tambah user
	protectedAdmin.GET("/users", controllers.GetAllUsers)       // Dapatkan semua user
	protectedAdmin.PUT("/users/:user_id", controllers.UpdateUser) // Update user
	protectedAdmin.DELETE("/users/:user_id", controllers.DeleteUser) // Hapus user

	// Routes untuk Device Management (Hanya Admin)
	protectedAdmin.POST("/devices", controllers.CreateDeviceAdmin)   // Tambah device
	protectedAdmin.GET("/devices", controllers.GetAllDevicesAdmin)   // Dapatkan semua device
	protectedAdmin.PUT("/devices/:device_id", controllers.UpdateDeviceAdmin) // Update device
	protectedAdmin.DELETE("/devices/:device_id", controllers.DeleteDevice) // Hapus device

	// Routes untuk Sensor Data Management (Hanya Admin)
	protectedAdmin.GET("/sensors/:device_id", controllers.GetSensorData) // Ambil data sensor dari device tertentu
	protectedAdmin.DELETE("/sensors/:sensor_id", controllers.DeleteSensorData) // Hapus data sensor tertentu
	return r
}
