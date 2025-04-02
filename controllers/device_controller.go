package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"backend/models"
	"backend/config"

	"github.com/gin-gonic/gin"
)

// GenerateAPIKey - Membuat API Key unik untuk perangkat
func GenerateAPIKey() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateDevice - Menambahkan device baru untuk user
func CreateDevice(c *gin.Context) {
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


func GetAllDevices(c *gin.Context) {
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


func UpdateDevice(c *gin.Context) {
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

// GetDeviceStatus - Mendapatkan status device berdasarkan API Key
func GetDeviceStatus(c *gin.Context) {
	// Mengambil device_id dari context setelah middleware APIKeyMiddleware
	deviceID, exists := c.Get("device_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var device models.Device
	if err := database.DB.Where("id = ?", deviceID).First(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch device"})
		return
	}

	// Mengembalikan data delay dan current_state
	c.JSON(http.StatusOK, gin.H{
		"delay":        device.Delay,
		"current_state": device.CurrentState,
	})
}


// UpdateDeviceState - Mengupdate status device (misalnya aktif/inaktif)
// func UpdateDevice(c *gin.Context) {
// 	// Ambil ID perangkat dari parameter URL
// 	deviceID, err := strconv.Atoi(c.Param("device_id")) // Sesuai dengan route
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
// 		return
// 	}


// 	// Ambil user ID dari token JWT
// 	userID, exists := c.Get("user_id")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 		return
// 	}

// 	// Konversi user ID ke uint
// 	userIDUint, ok := userID.(uint)
// 	if !ok {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID type error"})
// 		return
// 	}

// 	// Cari device berdasarkan ID dan user ID (agar user hanya bisa edit device miliknya)
// 	var device models.Device
// 	if err := database.DB.Where("id = ? AND user_id = ?", deviceID, userIDUint).First(&device).Error; err != nil {
// 		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to edit this device"})
// 		return
// 	}

// 	// Ambil data yang dikirimkan dalam body request
// 	var input struct {
// 		CurrentState string `json:"current_state"`
// 		Delay        int    `json:"delay"`
// 	}

// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
// 		return
// 	}

// 	// Update device
// 	device.CurrentState = input.CurrentState
// 	device.Delay = input.Delay

// 	if err := database.DB.Save(&device).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Device updated successfully"})
// }



