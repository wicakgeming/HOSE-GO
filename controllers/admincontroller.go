package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"golang.org/x/crypto/bcrypt"
	"backend/models"
	"backend/config"
	"time"

	"github.com/gin-gonic/gin"
)

// GenerateAPIKey - Membuat API Key unik untuk perangkat
func GenerateAPIKeyAdmin() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateUser - Menambahkan user baru oleh admin
func CreateUser(c *gin.Context) {
	var input struct {
		Username       string  `json:"username"`
		Password       string  `json:"password"`
		Email          string  `json:"email"`
		Role           string  `json:"role"`
		FullName       *string `json:"full_name"`
		DateOfBirth    *string `json:"date_of_birth"`
		MedicalHistory *string `json:"medical_history"`
		Address        *string `json:"address"`
		Province       *string `json:"province"`
		City           *string `json:"city"`
		PostalCode     *string `json:"postal_code"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Parse DateOfBirth jika tidak null
	var parsedDate *time.Time
	if input.DateOfBirth != nil && *input.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", *input.DateOfBirth) // Format YYYY-MM-DD
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format (YYYY-MM-DD required)"})
			return
		}
		parsedDate = &parsed
	}

	// Buat user baru
	user := models.User{
		Username:       input.Username,
		Password:       string(hashedPassword),
		Email:          input.Email,
		Role:           input.Role,
		FullName:       input.FullName,
		DateOfBirth:    parsedDate,
		MedicalHistory: input.MedicalHistory,
		Address:        input.Address,
		Province:       input.Province,
		City:           input.City,
		PostalCode:     input.PostalCode,
	}

	// Simpan ke database
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

// GetAllUsers - Mendapatkan semua user
func GetAllUsers(c *gin.Context) {
	var users []models.User
	if err := database.DB.Select("id, username, email, role, full_name, date_of_birth, medical_history, address, province, city, postal_code, email_verified, created_at, updated_at").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func UpdateUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var input struct {
		Username       *string `json:"username"`
		Email          *string `json:"email"`
		Role           *string `json:"role"`
		FullName       *string `json:"full_name"`
		DateOfBirth    *string `json:"date_of_birth"`
		MedicalHistory *string `json:"medical_history"`
		Address        *string `json:"address"`
		Province       *string `json:"province"`
		City           *string `json:"city"`
		PostalCode     *string `json:"postal_code"`
		Password       *string `json:"password,omitempty"` // Opsional, tidak harus dikirim
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Perbarui hanya jika ada perubahan
	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Role != nil {
		user.Role = *input.Role
	}
	if input.FullName != nil {
		user.FullName = input.FullName
	}
	if input.MedicalHistory != nil {
		user.MedicalHistory = input.MedicalHistory
	}
	if input.Address != nil {
		user.Address = input.Address
	}
	if input.Province != nil {
		user.Province = input.Province
	}
	if input.City != nil {
		user.City = input.City
	}
	if input.PostalCode != nil {
		user.PostalCode = input.PostalCode
	}

	// Parsing DateOfBirth jika diberikan
	if input.DateOfBirth != nil && *input.DateOfBirth != "" {
		parsedDate, err := time.Parse("2006-01-02", *input.DateOfBirth) // Format YYYY-MM-DD
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format (YYYY-MM-DD required)"})
			return
		}
		user.DateOfBirth = &parsedDate
	}

	// Jika password diisi, hash password baru
	if input.Password != nil && *input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt password"})
			return
		}
		user.Password = string(hashedPassword)
	}

	// Simpan perubahan ke database
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func DeleteUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := database.DB.Delete(&models.User{}, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}


// CreateDevice - Menambahkan device baru untuk user
func CreateDeviceAdmin(c *gin.Context) {
	// Pastikan hanya admin yang bisa akses
	role, _ := c.Get("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can add devices"})
		return
	}

	var device models.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate API Key untuk device baru
	device.APIKey = GenerateAPIKey()

	// Simpan ke database
	if err := database.DB.Create(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device created successfully", "api_key": device.APIKey})
}

func GetAllDevicesAdmin(c *gin.Context) {
	// Pastikan hanya admin yang bisa akses
	role, _ := c.Get("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can view all devices"})
		return
	}

	var devices []models.Device
	if err := database.DB.Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve devices"})
		return
	}

	c.JSON(http.StatusOK, devices)
}

func UpdateDeviceAdmin(c *gin.Context) {
	// Ambil ID perangkat dari parameter URL dan konversi ke uint
	deviceID, err := strconv.ParseUint(c.Param("device_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	// Ambil user ID dan role dari token JWT
	userID, exists := c.Get("user_id")
	role, _ := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Konversi user ID ke uint
	userIDUint := userID.(uint)

	// Cari device berdasarkan ID
	var device models.Device
	if role != "admin" {
		// Jika bukan admin, pastikan device milik user yang sedang login
		if err := database.DB.Where("id = ? AND user_id = ?", uint(deviceID), userIDUint).First(&device).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to edit this device"})
			return
		}
	} else {
		// Jika admin, tidak perlu cek user_id
		if err := database.DB.Where("id = ?", uint(deviceID)).First(&device).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
			return
		}
	}

	// Ambil data yang dikirimkan dalam body request
	var input struct {
		CurrentState string `json:"current_state"`
		Delay        int    `json:"delay"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Update device
	device.CurrentState = input.CurrentState
	device.Delay = input.Delay

	if err := database.DB.Save(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device updated successfully"})
}

// DeleteDevice - Menghapus device berdasarkan ID
func DeleteDevice(c *gin.Context) {
	// Ambil ID perangkat dari parameter URL dan konversi ke uint
	deviceID, err := strconv.Atoi(c.Param("device_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	// Ambil user ID dan role dari token JWT
	userID, exists := c.Get("user_id")
	role, _ := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Konversi user ID ke uint
	userIDUint := userID.(uint)

	// Cari device berdasarkan ID
	var device models.Device
	if role != "admin" {
		// Jika bukan admin, pastikan device milik user yang sedang login
		if err := database.DB.Where("id = ? AND user_id = ?", deviceID, userIDUint).First(&device).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to delete this device"})
			return
		}
	} else {
		// Jika admin, tidak perlu cek user_id
		if err := database.DB.Where("id = ?", deviceID).First(&device).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
			return
		}
	}

	// Hapus device
	if err := database.DB.Delete(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device deleted successfully"})
}

// DeleteSensorData - Menghapus data sensor berdasarkan ID
func DeleteSensorData(c *gin.Context) {
	// Ambil ID sensor dari parameter URL dan konversi ke uint
	sensorID, err := strconv.Atoi(c.Param("sensor_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}

	// Ambil user ID dan role dari token JWT
	userID, exists := c.Get("user_id")
	role, _ := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Konversi user ID ke uint
	userIDUint := userID.(uint)

	// Cari data sensor berdasarkan ID sensor
	var sensorData models.SensorData
	if err := database.DB.First(&sensorData, sensorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor data not found"})
		return
	}

	// Jika bukan admin, pastikan data sensor yang akan dihapus milik perangkat pengguna yang sedang login
	if role != "admin" {
		var device models.Device
		// Cek apakah perangkat yang terkait dengan data sensor milik user yang sedang login
		if err := database.DB.Where("id = ? AND user_id = ?", sensorData.DeviceID, userIDUint).First(&device).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to delete this sensor data"})
			return
		}
	}

	// Hapus data sensor
	if err := database.DB.Delete(&sensorData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sensor data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sensor data deleted successfully"})
}


