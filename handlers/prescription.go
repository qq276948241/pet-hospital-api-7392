package handlers

import (
	"errors"
	"pet-hospital/models"
	"pet-hospital/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BusinessError struct {
	Code    int
	Message string
}

func (e *BusinessError) Error() string {
	return e.Message
}

type PrescriptionItemRequest struct {
	MedicineID uint   `json:"medicine_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required"`
	Dosage     string `json:"dosage"`
}

type PrescriptionRequest struct {
	MedicalRecordID uint                       `json:"medical_record_id" binding:"required"`
	DoctorID        uint                       `json:"doctor_id" binding:"required"`
	PetID           uint                       `json:"pet_id" binding:"required"`
	Advice          string                     `json:"advice"`
	Items           []PrescriptionItemRequest `json:"items" binding:"required"`
}

func CreatePrescription(c *gin.Context) {
	var req PrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var record models.MedicalRecord
	if err := models.DB.First(&record, req.MedicalRecordID).Error; err != nil {
		utils.NotFound(c, "就诊记录不存在")
		return
	}

	err := models.DB.Transaction(func(tx *gorm.DB) error {
		prescription := models.Prescription{
			MedicalRecordID: req.MedicalRecordID,
			DoctorID:        req.DoctorID,
			PetID:           req.PetID,
			Advice:          req.Advice,
			Status:          "pending",
		}

		if err := tx.Create(&prescription).Error; err != nil {
			return err
		}

		for _, item := range req.Items {
			var medicine models.Medicine
			if err := tx.First(&medicine, item.MedicineID).Error; err != nil {
				return err
			}

			if medicine.Stock < item.Quantity {
				return &BusinessError{Code: 400, Message: medicine.Name + " 库存不足"}
			}

			prescriptionItem := models.PrescriptionItem{
				PrescriptionID: prescription.ID,
				MedicineID:     item.MedicineID,
				MedicineName:   medicine.Name,
				Quantity:       item.Quantity,
				UnitPrice:      medicine.Price,
				Dosage:         item.Dosage,
			}

			if err := tx.Create(&prescriptionItem).Error; err != nil {
				return err
			}

			medicine.Stock -= item.Quantity
			if err := tx.Save(&medicine).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		var be *BusinessError
		if errors.As(err, &be) {
			utils.Error(c, be.Code, be.Message)
			return
		}
		utils.InternalError(c, "开处方失败: "+err.Error())
		return
	}

	var prescription models.Prescription
	models.DB.Preload("Items").Preload("Items.Medicine").
		Where("medical_record_id = ?", req.MedicalRecordID).
		Order("created_at DESC").First(&prescription)

	utils.Success(c, prescription)
}

func GetPrescriptions(c *gin.Context) {
	userID := c.GetUint("user_id")
	petID := c.Query("pet_id")
	status := c.Query("status")

	var prescriptions []models.Prescription
	query := models.DB.Preload("Items").Preload("Items.Medicine").Preload("Doctor")

	if petID != "" {
		query = query.Where("pet_id = ?", petID)
	} else {
		var pets []models.Pet
		models.DB.Where("owner_id = ?", userID).Find(&pets)
		petIDs := make([]uint, len(pets))
		for i, pet := range pets {
			petIDs[i] = pet.ID
		}
		if len(petIDs) > 0 {
			query = query.Where("pet_id IN ?", petIDs)
		} else {
			query = query.Where("1 = 0")
		}
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&prescriptions).Error; err != nil {
		utils.InternalError(c, "获取处方列表失败: "+err.Error())
		return
	}

	utils.Success(c, prescriptions)
}

func GetPrescriptionByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	prescriptionID, _ := strconv.Atoi(c.Param("id"))

	var prescription models.Prescription
	if err := models.DB.Preload("Items").Preload("Items.Medicine").Preload("Doctor").
		First(&prescription, prescriptionID).Error; err != nil {
		utils.NotFound(c, "处方不存在")
		return
	}

	var pet models.Pet
	if err := models.DB.Where("id = ? AND owner_id = ?", prescription.PetID, userID).First(&pet).Error; err != nil {
		utils.Forbidden(c, "无权访问该处方")
		return
	}

	utils.Success(c, prescription)
}

func GetPrescriptionsByPet(c *gin.Context) {
	userID := c.GetUint("user_id")
	petID, _ := strconv.Atoi(c.Param("pet_id"))

	var pet models.Pet
	if err := models.DB.Where("id = ? AND owner_id = ?", petID, userID).First(&pet).Error; err != nil {
		utils.NotFound(c, "宠物不存在")
		return
	}

	var prescriptions []models.Prescription
	if err := models.DB.Preload("Items").Preload("Items.Medicine").Preload("Doctor").
		Where("pet_id = ?", petID).Order("created_at DESC").Find(&prescriptions).Error; err != nil {
		utils.InternalError(c, "获取处方列表失败: "+err.Error())
		return
	}

	utils.Success(c, prescriptions)
}

func UpdatePrescriptionStatus(c *gin.Context) {
	prescriptionID, _ := strconv.Atoi(c.Param("id"))

	var prescription models.Prescription
	if err := models.DB.First(&prescription, prescriptionID).Error; err != nil {
		utils.NotFound(c, "处方不存在")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	validStatuses := map[string]bool{
		"pending": true,
		"paid":    true,
		"used":    true,
	}

	if !validStatuses[req.Status] {
		utils.BadRequest(c, "无效的状态")
		return
	}

	prescription.Status = req.Status

	if err := models.DB.Save(&prescription).Error; err != nil {
		utils.InternalError(c, "更新状态失败: "+err.Error())
		return
	}

	utils.Success(c, prescription)
}
