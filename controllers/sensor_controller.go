package controllers

import (
	"net/http"
	"backend/models"
	"backend/config"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AddSensorData - ESP32 mengirim data sensor ke API
func AddSensorData(c *gin.Context) {
	// Ambil device_id dari context (sudah divalidasi di middleware)
	deviceID, exists := c.Get("device_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		BPM  float64 `json:"bpm" binding:"required"`
		SpO2 float64 `json:"spo2" binding:"required"`
		Temp float64 `json:"temp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simpan data sensor dengan device_id dari context
	sensorData := models.SensorData{
		DeviceID:  deviceID.(uint),
		BPM:       input.BPM,
		SpO2:      input.SpO2,
		Temp:      input.Temp,
		Timestamp: time.Now(),
	}

	if err := database.DB.Create(&sensorData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add sensor data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sensor data added successfully"})
}


// GetSensorData - Mendapatkan data sensor dari device tertentu
func GetSensorData(c *gin.Context) {
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


