package models

import (
	"time"
)

type Notification struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"not null;index" json:"user_id"`
	AppointmentID   uint      `gorm:"index" json:"appointment_id"`
	Type            string    `gorm:"size:30;not null" json:"type"`
	Title           string    `gorm:"size:100;not null" json:"title"`
	Content         string    `gorm:"size:1000;not null" json:"content"`
	ScheduledTime   time.Time `gorm:"index" json:"scheduled_time"`
	SentTime        *time.Time `json:"sent_time"`
	IsSent          bool      `gorm:"default:false;index" json:"is_sent"`
	IsRead          bool      `gorm:"default:false;index" json:"is_read"`
	ReadTime        *time.Time `json:"read_time"`
	Channel         string    `gorm:"size:20;default:'in_app'" json:"channel"`
	Remark          string    `gorm:"size:255" json:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Appointment     *Appointment `gorm:"foreignKey:AppointmentID" json:"appointment,omitempty"`
	User            *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
