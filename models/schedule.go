package models

import (
	"time"
)

type Schedule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DoctorID  uint      `gorm:"not null;index" json:"doctor_id"`
	Weekday   string    `gorm:"size:10;not null" json:"weekday"`
	Shift     string    `gorm:"size:10;not null" json:"shift"`
	MaxSlots  int       `gorm:"default:10" json:"max_slots"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Doctor    *Doctor   `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
}
