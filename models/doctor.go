package models

import (
	"time"
)

type Doctor struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Name           string     `gorm:"size:50;not null" json:"name"`
	Specialization string     `gorm:"size:50" json:"specialization"`
	Title          string     `gorm:"size:50" json:"title"`
	Phone          string     `gorm:"size:20" json:"phone"`
	Introduction   string     `gorm:"size:500" json:"introduction"`
	Avatar         string     `gorm:"size:255" json:"avatar"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Schedules      []Schedule `gorm:"foreignKey:DoctorID" json:"schedules,omitempty"`
}
