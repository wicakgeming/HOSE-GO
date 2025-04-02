package controllers

import (
	"net/http"
	"backend/models"
	"backend/config"
	"time"

	"github.com/gin-gonic/gin"
)

// AddSensorData - ESP32 mengirim data sensor ke API
func AddSensorData(c *gin.Context) {
	var input struct {
		DeviceID uint    `json:"device_id" binding:"required"`
		BPM      float64 `json:"bpm" binding:"required"`
		SpO2     float64 `json:"spo2" binding:"required"`
		Temp     float64 `json:"temp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sensorData := models.SensorData{
		DeviceID:  input.DeviceID,
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
	// Ambil user_id dari token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Ambil device_id dari parameter URL
	deviceID := c.Param("device_id")

	// Cek apakah device dengan device_id ini dimiliki oleh user yang sedang login
	var device models.Device
	if err := database.DB.Where("id = ? AND user_id = ?", deviceID, userID).First(&device).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to access this device's sensor data"})
		return
	}

	// Jika device milik user, ambil data sensor
	var sensorData []models.SensorData
	if err := database.DB.Where("device_id = ?", deviceID).Order("timestamp desc").Limit(10).Find(&sensorData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sensor data"})
		return
	}

	c.JSON(http.StatusOK, sensorData)
}

