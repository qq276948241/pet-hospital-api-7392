package models

import (
	"time"
)

type MedicalRecord struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	AppointmentID uint           `gorm:"not null;index" json:"appointment_id"`
	PetID         uint           `gorm:"not null;index" json:"pet_id"`
	DoctorID      uint           `gorm:"not null;index" json:"doctor_id"`
	Diagnosis     string         `gorm:"size:1000;not null" json:"diagnosis" binding:"required"`
	Symptoms      string         `gorm:"size:1000" json:"symptoms"`
	Treatment     string         `gorm:"size:1000" json:"treatment"`
	Notes         string         `gorm:"size:1000" json:"notes"`
	Temperature   float64        `json:"temperature"`
	HeartRate     int            `json:"heart_rate"`
	BreathRate    int            `json:"breath_rate"`
	Weight        float64        `json:"weight"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Appointment   *Appointment   `gorm:"foreignKey:AppointmentID" json:"appointment,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID" json:"pet,omitempty"`
	Doctor        *Doctor        `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	Prescription  *Prescription `gorm:"foreignKey:MedicalRecordID" json:"prescription,omitempty"`
}
