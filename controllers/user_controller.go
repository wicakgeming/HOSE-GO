package controllers

import (
	database "backend/config"
	"backend/models"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GenerateAPIKey - Membuat API Key unik untuk perangkat
func GenerateAPIKey() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// =================== Device Management ===================

// GetDevicesByUser - Mendapatkan semua device milik user tertentu
func GetDevicesByUser(c *gin.Context) {
	// Ambil user_id dari token JWT
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var devices []models.Device
	if err := database.DB.Where("user_id = ?", userID).Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch devices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"devices": devices})
}

// GetDeviceByUser - Mendapatkan device tertentu milik user
func UpdateDeviceByUser(c *gin.Context) {
	// Ambil ID perangkat dari parameter URL dan konversi ke uint
	deviceID, err := strconv.ParseUint(c.Param("device_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	// Ambil user ID dari token JWT
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Konversi user ID ke uint
	userIDUint := userID.(uint)

	// Cari device berdasarkan ID dan user ID (agar user hanya bisa edit device miliknya)
	var device models.Device
	if err := database.DB.Where("id = ? AND user_id = ?", uint(deviceID), userIDUint).First(&device).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to edit this device"})
		return
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

// AddDeviceByUser - Menambahkan device baru untuk user tertentu
func AddDeviceByUser(c *gin.Context) {
	// Get user_id from token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var device models.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set device owner and generate API key
	device.UserID = userID.(uint)
	device.APIKey = GenerateAPIKey()

	if err := database.DB.Create(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Device added successfully",
		"device":  device,
		"api_key": device.APIKey,
	})
}

// DeleteDeviceByUser - Menghapus device tertentu yang dimiliki user
func DeleteDeviceByUser(c *gin.Context) {
	// Get user_id from token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get device ID from URL
	deviceID, err := strconv.ParseUint(c.Param("device_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	// Find the device and ensure ownership
	var device models.Device
	if err := database.DB.Where("id = ? AND user_id = ?", uint(deviceID), userID.(uint)).First(&device).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to delete this device"})
		return
	}

	if err := database.DB.Delete(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device deleted successfully"})
}

// GetSensorData - Mendapatkan data sensor dari device tertentu
func GetSensorDataByUser(c *gin.Context) {
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

	// Jika bukan admin, pastikan data sensor yang diambil adalah milik user yang sedang login
	if role != "admin" {
		var device models.Device
		// Cek apakah perangkat milik user yang sedang login
		if err := database.DB.Where("id = ? AND user_id = ?", deviceID, userIDUint).First(&device).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to access this device's sensor data"})
			return
		}
	}

	// Ambil data sensor berdasarkan device ID
	var sensorData []models.SensorData
	if err := database.DB.Where("device_id = ?", deviceID).Find(&sensorData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sensor data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sensor_data": sensorData})
}

// =================== User Management ===================

// UserInfoByUser - Mendapatkan informasi user
func UserInfoByUser(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Password sudah tidak di-serialize karena di model json:"-"
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// DeleteUserByUser - Menghapus user
func DeleteUserByUser(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	if err := database.DB.Delete(&models.User{}, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// UpdateUserByUser - Mengubah informasi user
func UpdateUserByUser(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var input struct {
		Username       string     `json:"username" binding:"required"`
		Email          string     `json:"email" binding:"required,email"`
		FullName       *string    `json:"full_name"`
		DateOfBirth    *time.Time `json:"date_of_birth"`
		MedicalHistory *string    `json:"medical_history"`
		Address        *string    `json:"address"`
		Province       *string    `json:"province"`
		City           *string    `json:"city"`
		PostalCode     *string    `json:"postal_code"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "detail": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Username = input.Username
	user.Email = input.Email
	user.FullName = input.FullName
	user.DateOfBirth = input.DateOfBirth
	user.MedicalHistory = input.MedicalHistory
	user.Address = input.Address
	user.Province = input.Province
	user.City = input.City
	user.PostalCode = input.PostalCode

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// ChangePasswordByUser - Mengubah password user
func ChangePasswordByUser(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "detail": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Cek apakah password lama cocok
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Old password is incorrect"})
		return
	}

	// Hash password baru
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	user.Password = string(hashedPassword)
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
