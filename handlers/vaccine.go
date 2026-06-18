package handlers

import (
	"pet-hospital/models"
	"pet-hospital/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type VaccineRecordRequest struct {
	PetID           uint   `json:"pet_id" binding:"required"`
	VaccineID       uint   `json:"vaccine_id" binding:"required"`
	VaccinationDate string `json:"vaccination_date" binding:"required"`
	DoctorID        uint   `json:"doctor_id"`
	BatchNumber     string `json:"batch_number"`
	Remark          string `json:"remark"`
}

func GetVaccines(c *gin.Context) {
	vaccineType := c.Query("type")

	var vaccines []models.Vaccine
	query := models.DB

	if vaccineType != "" {
		query = query.Where("type = ?", vaccineType)
	}

	if err := query.Find(&vaccines).Error; err != nil {
		utils.InternalError(c, "获取疫苗列表失败: "+err.Error())
		return
	}

	utils.Success(c, vaccines)
}

func GetVaccineByID(c *gin.Context) {
	vaccineID, _ := strconv.Atoi(c.Param("id"))

	var vaccine models.Vaccine
	if err := models.DB.First(&vaccine, vaccineID).Error; err != nil {
		utils.NotFound(c, "疫苗不存在")
		return
	}

	utils.Success(c, vaccine)
}

func AddVaccineRecord(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req VaccineRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var pet models.Pet
	if err := models.DB.Where("id = ? AND owner_id = ?", req.PetID, userID).First(&pet).Error; err != nil {
		utils.NotFound(c, "宠物不存在")
		return
	}

	var vaccine models.Vaccine
	if err := models.DB.First(&vaccine, req.VaccineID).Error; err != nil {
		utils.NotFound(c, "疫苗不存在")
		return
	}

	nextDueDate := ""
	if vaccine.DurationDays > 0 {
		t, err := time.Parse("2006-01-02", req.VaccinationDate)
		if err == nil {
			nextDueDate = t.AddDate(0, 0, vaccine.DurationDays).Format("2006-01-02")
		}
	}

	record := models.VaccineRecord{
		PetID:           req.PetID,
		VaccineID:       req.VaccineID,
		VaccineName:     vaccine.Name,
		VaccinationDate: req.VaccinationDate,
		NextDueDate:     nextDueDate,
		DoctorID:        req.DoctorID,
		BatchNumber:     req.BatchNumber,
		Remark:          req.Remark,
	}

	if err := models.DB.Create(&record).Error; err != nil {
		utils.InternalError(c, "添加接种记录失败: "+err.Error())
		return
	}

	utils.Success(c, record)
}

func GetVaccineRecords(c *gin.Context) {
	userID := c.GetUint("user_id")
	petID := c.Query("pet_id")

	var records []models.VaccineRecord
	query := models.DB.Preload("Vaccine")

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

	if err := query.Order("vaccination_date DESC").Find(&records).Error; err != nil {
		utils.InternalError(c, "获取接种记录失败: "+err.Error())
		return
	}

	utils.Success(c, records)
}

func GetVaccineRecordsByPet(c *gin.Context) {
	userID := c.GetUint("user_id")
	petID, _ := strconv.Atoi(c.Param("pet_id"))

	var pet models.Pet
	if err := models.DB.Where("id = ? AND owner_id = ?", petID, userID).First(&pet).Error; err != nil {
		utils.NotFound(c, "宠物不存在")
		return
	}

	var records []models.VaccineRecord
	if err := models.DB.Preload("Vaccine").Where("pet_id = ?", petID).
		Order("vaccination_date DESC").Find(&records).Error; err != nil {
		utils.InternalError(c, "获取接种记录失败: "+err.Error())
		return
	}

	utils.Success(c, records)
}

func GetVaccineReminders(c *gin.Context) {
	userID := c.GetUint("user_id")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	var pets []models.Pet
	models.DB.Where("owner_id = ?", userID).Find(&pets)
	petIDs := make([]uint, len(pets))
	for i, pet := range pets {
		petIDs[i] = pet.ID
	}

	if len(petIDs) == 0 {
		utils.Success(c, []interface{}{})
		return
	}

	today := time.Now()
	reminderDate := today.AddDate(0, 0, days).Format("2006-01-02")
	todayStr := today.Format("2006-01-02")

	var records []models.VaccineRecord
	if err := models.DB.Preload("Vaccine").Preload("Pet").
		Where("pet_id IN ? AND next_due_date >= ? AND next_due_date <= ?",
			petIDs, todayStr, reminderDate).
		Order("next_due_date ASC").Find(&records).Error; err != nil {
		utils.InternalError(c, "获取疫苗提醒失败: "+err.Error())
		return
	}

	reminders := make([]gin.H, 0)
	for _, record := range records {
		dueDate, _ := time.Parse("2006-01-02", record.NextDueDate)
		daysLeft := int(time.Until(dueDate).Hours() / 24)

		status := "upcoming"
		if daysLeft < 0 {
			status = "overdue"
		} else if daysLeft <= 7 {
			status = "urgent"
		}

		reminders = append(reminders, gin.H{
			"id":              record.ID,
			"pet_id":          record.PetID,
			"pet_name":        record.Pet.Name,
			"vaccine_id":      record.VaccineID,
			"vaccine_name":    record.VaccineName,
			"vaccination_date": record.VaccinationDate,
			"next_due_date":   record.NextDueDate,
			"days_left":       daysLeft,
			"status":          status,
			"description":     record.Vaccine.Description,
		})
	}

	utils.Success(c, reminders)
}

func GetUpcomingReminders(c *gin.Context) {
	userID := c.GetUint("user_id")

	var pets []models.Pet
	models.DB.Where("owner_id = ?", userID).Find(&pets)
	petIDs := make([]uint, len(pets))
	for i, pet := range pets {
		petIDs[i] = pet.ID
	}

	if len(petIDs) == 0 {
		utils.Success(c, gin.H{
			"overdue":  0,
			"urgent":   0,
			"upcoming": 0,
			"list":     []interface{}{},
		})
		return
	}

	today := time.Now()
	weekLater := today.AddDate(0, 0, 7).Format("2006-01-02")
	monthLater := today.AddDate(0, 0, 30).Format("2006-01-02")
	todayStr := today.Format("2006-01-02")

	var allRecords []models.VaccineRecord
	models.DB.Preload("Vaccine").Preload("Pet").
		Where("pet_id IN ? AND next_due_date >= ? AND next_due_date <= ?", petIDs, todayStr, monthLater).
		Order("next_due_date ASC").Find(&allRecords)

	overdueCount := 0
	urgentCount := 0
	upcomingCount := 0

	reminders := make([]gin.H, 0)
	for _, record := range allRecords {
		dueDate, _ := time.Parse("2006-01-02", record.NextDueDate)
		daysLeft := int(time.Until(dueDate).Hours() / 24)

		status := "upcoming"
		if daysLeft < 0 {
			status = "overdue"
			overdueCount++
		} else if daysLeft <= 7 {
			status = "urgent"
			urgentCount++
		} else {
			upcomingCount++
		}

		if record.NextDueDate <= weekLater {
			reminders = append(reminders, gin.H{
				"id":              record.ID,
				"pet_id":          record.PetID,
				"pet_name":        record.Pet.Name,
				"vaccine_name":    record.VaccineName,
				"next_due_date":   record.NextDueDate,
				"days_left":       daysLeft,
				"status":          status,
			})
		}
	}

	utils.Success(c, gin.H{
		"overdue":  overdueCount,
		"urgent":   urgentCount,
		"upcoming": upcomingCount,
		"list":     reminders,
	})
}
