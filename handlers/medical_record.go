package handlers

import (
	"pet-hospital/models"
	"pet-hospital/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MedicalRecordRequest struct {
	AppointmentID uint    `json:"appointment_id" binding:"required"`
	PetID         uint    `json:"pet_id" binding:"required"`
	DoctorID      uint    `json:"doctor_id" binding:"required"`
	Diagnosis     string  `json:"diagnosis" binding:"required"`
	Symptoms      string  `json:"symptoms"`
	Treatment     string  `json:"treatment"`
	Notes         string  `json:"notes"`
	Temperature   float64 `json:"temperature"`
	HeartRate     int     `json:"heart_rate"`
	BreathRate    int     `json:"breath_rate"`
	Weight        float64 `json:"weight"`
}

func CreateMedicalRecord(c *gin.Context) {
	var req MedicalRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var appointment models.Appointment
	if err := models.DB.First(&appointment, req.AppointmentID).Error; err != nil {
		utils.NotFound(c, "预约记录不存在")
		return
	}

	record := models.MedicalRecord{
		AppointmentID: req.AppointmentID,
		PetID:         req.PetID,
		DoctorID:      req.DoctorID,
		Diagnosis:     req.Diagnosis,
		Symptoms:      req.Symptoms,
		Treatment:     req.Treatment,
		Notes:         req.Notes,
		Temperature:   req.Temperature,
		HeartRate:     req.HeartRate,
		BreathRate:    req.BreathRate,
		Weight:        req.Weight,
	}

	if err := models.DB.Create(&record).Error; err != nil {
		utils.InternalError(c, "创建就诊记录失败: "+err.Error())
		return
	}

	appointment.Status = "completed"
	models.DB.Save(&appointment)

	utils.Success(c, record)
}

func GetMedicalRecords(c *gin.Context) {
	userID := c.GetUint("user_id")
	petID := c.Query("pet_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var records []models.MedicalRecord
	var total int64

	query := models.DB.Preload("Doctor").Preload("Prescription").Preload("Prescription.Items")

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

	query.Model(&models.MedicalRecord{}).Count(&total)
	offset := (page - 1) * pageSize

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records).Error; err != nil {
		utils.InternalError(c, "获取就诊记录失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":       records,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

func GetMedicalRecordByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	recordID, _ := strconv.Atoi(c.Param("id"))

	var record models.MedicalRecord
	if err := models.DB.Preload("Doctor").Preload("Prescription").Preload("Prescription.Items").
		First(&record, recordID).Error; err != nil {
		utils.NotFound(c, "就诊记录不存在")
		return
	}

	var pet models.Pet
	if err := models.DB.Where("id = ? AND owner_id = ?", record.PetID, userID).First(&pet).Error; err != nil {
		utils.Forbidden(c, "无权访问该记录")
		return
	}

	utils.Success(c, record)
}

func GetMedicalRecordsByPet(c *gin.Context) {
	userID := c.GetUint("user_id")
	petID, _ := strconv.Atoi(c.Param("pet_id"))

	var pet models.Pet
	if err := models.DB.Where("id = ? AND owner_id = ?", petID, userID).First(&pet).Error; err != nil {
		utils.NotFound(c, "宠物不存在")
		return
	}

	var records []models.MedicalRecord
	if err := models.DB.Preload("Doctor").Preload("Prescription").Preload("Prescription.Items").
		Where("pet_id = ?", petID).Order("created_at DESC").Find(&records).Error; err != nil {
		utils.InternalError(c, "获取就诊记录失败: "+err.Error())
		return
	}

	utils.Success(c, records)
}

func UpdateMedicalRecord(c *gin.Context) {
	recordID, _ := strconv.Atoi(c.Param("id"))

	var record models.MedicalRecord
	if err := models.DB.First(&record, recordID).Error; err != nil {
		utils.NotFound(c, "就诊记录不存在")
		return
	}

	var req MedicalRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	record.Diagnosis = req.Diagnosis
	record.Symptoms = req.Symptoms
	record.Treatment = req.Treatment
	record.Notes = req.Notes
	record.Temperature = req.Temperature
	record.HeartRate = req.HeartRate
	record.BreathRate = req.BreathRate
	record.Weight = req.Weight

	if err := models.DB.Save(&record).Error; err != nil {
		utils.InternalError(c, "更新就诊记录失败: "+err.Error())
		return
	}

	utils.Success(c, record)
}
