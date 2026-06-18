package models

import (
	"time"
)

type Medicine struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Category    string    `gorm:"size:50" json:"category"`
	Price       float64   `json:"price"`
	Stock       int       `gorm:"default:0" json:"stock"`
	Unit        string    `gorm:"size:10" json:"unit"`
	Description string    `gorm:"size:500" json:"description"`
	Supplier    string    `gorm:"size:100" json:"supplier"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
