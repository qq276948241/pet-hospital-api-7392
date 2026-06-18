package models

import (
	"time"
)

type Appointment struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UserID          uint           `gorm:"not null;index" json:"user_id"`
	PetID           uint           `gorm:"not null;index" json:"pet_id"`
	DoctorID        uint           `gorm:"not null;index" json:"doctor_id"`
	AppointmentDate string         `gorm:"size:20;not null" json:"appointment_date"`
	Shift           string         `gorm:"size:10;not null" json:"shift"`
	SlotNumber      int            `json:"slot_number"`
	Status          string         `gorm:"size:20;default:'pending'" json:"status"`
	Symptoms        string         `gorm:"size:500" json:"symptoms"`
	Remark          string         `gorm:"size:500" json:"remark"`
	CancelReason    string         `gorm:"size:255" json:"cancel_reason"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	User            *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Pet             *Pet           `gorm:"foreignKey:PetID" json:"pet,omitempty"`
	Doctor          *Doctor        `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	MedicalRecord   *MedicalRecord `gorm:"foreignKey:AppointmentID" json:"medical_record,omitempty"`
}
