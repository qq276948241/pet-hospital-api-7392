package models

import (
	"time"
)

type Pet struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OwnerID     uint      `gorm:"not null;index" json:"owner_id"`
	Name        string    `gorm:"size:50;not null" json:"name" binding:"required"`
	Species     string    `gorm:"size:20;not null" json:"species" binding:"required"`
	Breed       string    `gorm:"size:50" json:"breed"`
	Gender      string    `gorm:"size:10" json:"gender"`
	Age         int       `json:"age"`
	Weight      float64   `json:"weight"`
	Color       string    `gorm:"size:30" json:"color"`
	BirthDate   string    `gorm:"size:20" json:"birth_date"`
	Neutered    bool      `json:"neutered"`
	Description string    `gorm:"size:500" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Owner       *User     `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
}
