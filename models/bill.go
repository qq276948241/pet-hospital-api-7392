package models

import (
	"time"
)

type Bill struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	AppointmentID   uint       `gorm:"index" json:"appointment_id"`
	UserID          uint       `gorm:"not null;index" json:"user_id"`
	PetID           uint       `gorm:"not null;index" json:"pet_id"`
	TotalAmount     float64    `gorm:"not null" json:"total_amount"`
	PaidAmount      float64    `gorm:"default:0" json:"paid_amount"`
	Status          string     `gorm:"size:20;default:'unpaid'" json:"status"`
	PaymentMethod   string     `gorm:"size:20" json:"payment_method"`
	PaymentTime     *time.Time `json:"payment_time"`
	Remark          string     `gorm:"size:500" json:"remark"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Appointment     *Appointment `gorm:"foreignKey:AppointmentID" json:"appointment,omitempty"`
	User            *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Pet             *Pet       `gorm:"foreignKey:PetID" json:"pet,omitempty"`
	Items           []BillItem `gorm:"foreignKey:BillID" json:"items,omitempty"`
}

type BillItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	BillID      uint      `gorm:"not null;index" json:"bill_id"`
	ItemType    string    `gorm:"size:50" json:"item_type"`
	ItemName    string    `gorm:"size:100" json:"item_name"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}
