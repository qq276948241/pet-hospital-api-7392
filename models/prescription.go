package models

import (
	"time"
)

type Prescription struct {
	ID              uint               `gorm:"primaryKey" json:"id"`
	MedicalRecordID uint               `gorm:"not null;index" json:"medical_record_id"`
	DoctorID        uint               `gorm:"not null;index" json:"doctor_id"`
	PetID           uint               `gorm:"not null;index" json:"pet_id"`
	Advice          string             `gorm:"size:1000" json:"advice"`
	Status          string             `gorm:"size:20;default:'pending'" json:"status"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	MedicalRecord   *MedicalRecord     `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Doctor          *Doctor            `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	Pet             *Pet               `gorm:"foreignKey:PetID" json:"pet,omitempty"`
	Items           []PrescriptionItem  `gorm:"foreignKey:PrescriptionID" json:"items,omitempty"`
}

type PrescriptionItem struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PrescriptionID uint      `gorm:"not null;index" json:"prescription_id"`
	MedicineID     uint      `gorm:"not null" json:"medicine_id"`
	MedicineName   string    `gorm:"size:100" json:"medicine_name"`
	Quantity       int       `gorm:"not null" json:"quantity"`
	UnitPrice      float64   `json:"unit_price"`
	Dosage         string    `gorm:"size:255" json:"dosage"`
	CreatedAt      time.Time `json:"created_at"`
	Medicine       *Medicine `gorm:"foreignKey:MedicineID" json:"medicine,omitempty"`
}
