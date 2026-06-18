package models

import (
	"time"
)

type Vaccine struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	Type         string    `gorm:"size:50" json:"type"`
	Price        float64   `json:"price"`
	DurationDays int       `json:"duration_days"`
	Description  string    `gorm:"size:500" json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type VaccineRecord struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	PetID           uint      `gorm:"not null;index" json:"pet_id"`
	VaccineID       uint      `gorm:"not null;index" json:"vaccine_id"`
	VaccineName     string    `gorm:"size:100" json:"vaccine_name"`
	VaccinationDate string    `gorm:"size:20;not null" json:"vaccination_date"`
	NextDueDate     string    `gorm:"size:20" json:"next_due_date"`
	DoctorID        uint      `json:"doctor_id"`
	BatchNumber     string    `gorm:"size:50" json:"batch_number"`
	Remark          string    `gorm:"size:255" json:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Pet             *Pet      `gorm:"foreignKey:PetID" json:"pet,omitempty"`
	Vaccine         *Vaccine `gorm:"foreignKey:VaccineID" json:"vaccine,omitempty"`
}
