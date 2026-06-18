package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50;not null" json:"username" binding:"required"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Phone     string    `gorm:"size:20;not null" json:"phone" binding:"required"`
	Email     string    `gorm:"size:100" json:"email"`
	RealName  string    `gorm:"size:50" json:"real_name"`
	Address   string    `gorm:"size:255" json:"address"`
	Role      string    `gorm:"size:20;default:'user'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Pets      []Pet     `gorm:"foreignKey:OwnerID" json:"pets,omitempty"`
}
