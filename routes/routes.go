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
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
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
		role, _ := c.Get("role")
		email, _ := c.Get("email")

		c.JSON(http.StatusOK, gin.H{
			"message":  "You are authorized",
			"user_id":  userID,
			"username": username,
			"role":     role,
			"email":    email,
		})
	})

	// User Routes (User)
	protected.GET("/user", controllers.UserInfoByUser)                       // Dapatkan informasi user
	protected.PATCH("/user", controllers.UpdateUserByUser)                   // Update informasi user
	protected.DELETE("/user", controllers.DeleteUserByUser)                  // Hapus user
	protected.PUT("/user/change-password", controllers.ChangePasswordByUser) // Ubah password user

	// Device Routes (User)
	protected.GET("/devices", controllers.GetDevicesByUser)                // Dapatkan semua device yang dimiliki user
	protected.PUT("/device/:device_id", controllers.UpdateDeviceByUser)    // Update device tertentu wajib dimiliki user
	protected.POST("/device", controllers.AddDeviceByUser)                 // Tambah device baru untuk user
	protected.DELETE("/device/:device_id", controllers.DeleteDeviceByUser) // Hapus device tertentu yang dimiliki user
	protected.GET("/sensor/:device_id", controllers.GetSensorDataByUser)   // Dapatkan data sensor dari device tertentu yang dimiliki user

	// =================== Device API Routes (Memerlukan API) ===================
	deviceAPI := r.Group("/api/device")
	deviceAPI.Use(middleware.APIKeyMiddleware())               // Middleware untuk memeriksa API Key
	deviceAPI.POST("/sensor", controllers.AddSensorDataByAPI)  // Endpoint untuk menambahkan data sensor ke device tertentu
	deviceAPI.GET("/status", controllers.GetDeviceStatusByAPI) // Endpoint untuk melihat status device

	// =================== Admin Routes (Memerlukan Token Admin) ===================
	protectedAdmin := r.Group("/admin")
	protectedAdmin.Use(middleware.AuthMiddleware(), middleware.AdminOnly())

	// Routes untuk User Management (Hanya Admin)
	protectedAdmin.POST("/users", controllers.CreateUserAdmin)            // Tambah user
	protectedAdmin.GET("/users", controllers.GetAllUsersAdmin)            // Dapatkan semua user
	protectedAdmin.PUT("/users/:user_id", controllers.UpdateUserAdmin)    // Update user
	protectedAdmin.DELETE("/users/:user_id", controllers.DeleteUserAdmin) // Hapus user

	// Routes untuk Device Management (Hanya Admin)
	protectedAdmin.POST("/devices", controllers.CreateDeviceAdmin)              // Tambah device
	protectedAdmin.GET("/devices", controllers.GetAllDevicesAdmin)              // Dapatkan semua device
	protectedAdmin.PUT("/devices/:device_id", controllers.UpdateDeviceAdmin)    // Update device
	protectedAdmin.DELETE("/devices/:device_id", controllers.DeleteDeviceAdmin) // Hapus device

	// Routes untuk Sensor Data Management (Hanya Admin)
	protectedAdmin.GET("/sensors/:device_id", controllers.GetSensorDataByAdmin)     // Ambil data sensor dari device tertentu
	protectedAdmin.DELETE("/sensors/:sensor_id", controllers.DeleteSensorDataAdmin) // Hapus data sensor tertentu
	return r
}
