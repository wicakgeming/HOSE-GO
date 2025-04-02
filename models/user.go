package models

import "time"

// Model User
type User struct {
    ID             uint      `gorm:"primaryKey" json:"id"`
    Username       string    `gorm:"unique;not null" json:"username"`
    Password       string    `gorm:"not null" json:"-"` 
    Email          string    `gorm:"unique;not null" json:"email"`
    Role           string    `gorm:"default:'user'" json:"role"`
    FullName       *string   `json:"full_name"`   
    DateOfBirth    *time.Time `json:"date_of_birth"`
    MedicalHistory *string   `json:"medical_history"`
    Address        *string   `json:"address"`
    Province       *string   `json:"province"`
    City           *string   `json:"city"`
    PostalCode     *string   `json:"postal_code"`
    EmailVerified  bool      `gorm:"default:false" json:"email_verified"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}


// Model Device (Alat yang dimiliki user)
type Device struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    UserID       uint      `gorm:"not null" json:"user_id"`
    User         User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"-"`
    Name         string    `gorm:"not null" json:"name"`
    APIKey       string    `gorm:"unique;not null" json:"api_key"` // API Key unik untuk ESP32-S3
    Delay        int       `gorm:"default:10" json:"delay"`
    CurrentState string    `gorm:"default:'inactive'" json:"current_state"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}


// Model SensorData (Data sensor dari alat)
type SensorData struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DeviceID  uint      `gorm:"not null" json:"device_id"`
	Device    Device    `gorm:"foreignKey:DeviceID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"-"`
	BPM       float64   `json:"bpm"`
	SpO2      float64   `json:"spo2"`
	Temp      float64   `json:"temp"`
	Timestamp time.Time `json:"timestamp"`
}
