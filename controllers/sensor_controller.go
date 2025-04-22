package controllers

import (
	"net/http"
	"backend/models"
	"backend/config"
	"time"
	"github.com/gin-gonic/gin"
)

// AddSensorData - ESP32 mengirim data sensor ke API
func AddSensorDataByAPI(c *gin.Context) {
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

func GetDeviceStatusByAPI(c *gin.Context) {
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


