package handlers

import (
	"pet-hospital/models"
	"pet-hospital/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetDoctors(c *gin.Context) {
	specialization := c.Query("specialization")

	var doctors []models.Doctor
	query := models.DB

	if specialization != "" {
		query = query.Where("specialization = ?", specialization)
	}

	if err := query.Find(&doctors).Error; err != nil {
		utils.InternalError(c, "获取医生列表失败: "+err.Error())
		return
	}

	utils.Success(c, doctors)
}

func GetDoctorByID(c *gin.Context) {
	doctorID, _ := strconv.Atoi(c.Param("id"))

	var doctor models.Doctor
	if err := models.DB.Preload("Schedules").First(&doctor, doctorID).Error; err != nil {
		utils.NotFound(c, "医生不存在")
		return
	}

	utils.Success(c, doctor)
}

func GetSchedules(c *gin.Context) {
	doctorID := c.Query("doctor_id")
	weekday := c.Query("weekday")

	var schedules []models.Schedule
	query := models.DB.Preload("Doctor")

	if doctorID != "" {
		query = query.Where("doctor_id = ?", doctorID)
	}
	if weekday != "" {
		query = query.Where("weekday = ?", weekday)
	}

	if err := query.Find(&schedules).Error; err != nil {
		utils.InternalError(c, "获取排班信息失败: "+err.Error())
		return
	}

	utils.Success(c, schedules)
}

func GetScheduleByDoctor(c *gin.Context) {
	doctorID, _ := strconv.Atoi(c.Param("id"))

	var schedules []models.Schedule
	if err := models.DB.Where("doctor_id = ?", doctorID).Find(&schedules).Error; err != nil {
		utils.InternalError(c, "获取医生排班失败: "+err.Error())
		return
	}

	utils.Success(c, schedules)
}
