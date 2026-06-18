package handlers

import (
	"fmt"
	"pet-hospital/models"
	"pet-hospital/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AppointmentRequest struct {
	PetID           uint   `json:"pet_id" binding:"required"`
	DoctorID        uint   `json:"doctor_id" binding:"required"`
	AppointmentDate string `json:"appointment_date" binding:"required"`
	Shift           string `json:"shift" binding:"required"`
	Symptoms        string `json:"symptoms"`
	Remark          string `json:"remark"`
}

func CreateAppointment(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req AppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var pet models.Pet
	if err := models.DB.Where("id = ? AND owner_id = ?", req.PetID, userID).First(&pet).Error; err != nil {
		utils.NotFound(c, "宠物不存在")
		return
	}

	var doctor models.Doctor
	if err := models.DB.First(&doctor, req.DoctorID).Error; err != nil {
		utils.NotFound(c, "医生不存在")
		return
	}

	t, err := time.Parse("2006-01-02", req.AppointmentDate)
	if err != nil {
		utils.BadRequest(c, "日期格式错误，请使用 YYYY-MM-DD 格式")
		return
	}

	weekdayMap := map[int]string{
		0: "周日", 1: "周一", 2: "周二", 3: "周三", 4: "周四", 5: "周五", 6: "周六",
	}
	weekday := weekdayMap[int(t.Weekday())]

	var schedule models.Schedule
	if err := models.DB.Where("doctor_id = ? AND weekday = ? AND shift = ?", req.DoctorID, weekday, req.Shift).First(&schedule).Error; err != nil {
		utils.BadRequest(c, "该医生该时段无排班")
		return
	}

	var bookedCount int64
	models.DB.Model(&models.Appointment{}).Where(
		"doctor_id = ? AND appointment_date = ? AND shift = ? AND status IN ?",
		req.DoctorID, req.AppointmentDate, req.Shift, []string{"pending", "confirmed"},
	).Count(&bookedCount)

	if bookedCount >= int64(schedule.MaxSlots) {
		utils.BadRequest(c, "该时段已约满")
		return
	}

	var existing models.Appointment
	if err := models.DB.Where(
		"user_id = ? AND pet_id = ? AND appointment_date = ? AND shift = ? AND status IN ?",
		userID, req.PetID, req.AppointmentDate, req.Shift, []string{"pending", "confirmed"},
	).First(&existing).Error; err == nil {
		utils.BadRequest(c, "您已在此时段预约过")
		return
	}

	appointment := models.Appointment{
		UserID:          userID,
		PetID:           req.PetID,
		DoctorID:        req.DoctorID,
		AppointmentDate: req.AppointmentDate,
		Shift:           req.Shift,
		SlotNumber:      int(bookedCount) + 1,
		Status:          "pending",
		Symptoms:        req.Symptoms,
		Remark:          req.Remark,
	}

	if err := models.DB.Create(&appointment).Error; err != nil {
		utils.InternalError(c, "预约失败: "+err.Error())
		return
	}

	notification, _ := utils.CreateAppointmentReminder(&appointment)

	utils.Success(c, gin.H{
		"appointment":  appointment,
		"notification": notification,
	})
}

func GetAppointments(c *gin.Context) {
	userID := c.GetUint("user_id")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var appointments []models.Appointment
	var total int64

	query := models.DB.Preload("Pet").Preload("Doctor").Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Model(&models.Appointment{}).Count(&total)
	offset := (page - 1) * pageSize

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&appointments).Error; err != nil {
		utils.InternalError(c, "获取预约列表失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":       appointments,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

func GetAppointmentByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	apptID, _ := strconv.Atoi(c.Param("id"))

	var appointment models.Appointment
	if err := models.DB.Preload("Pet").Preload("Doctor").Preload("MedicalRecord").
		Where("id = ? AND user_id = ?", apptID, userID).First(&appointment).Error; err != nil {
		utils.NotFound(c, "预约记录不存在")
		return
	}

	utils.Success(c, appointment)
}

func CancelAppointment(c *gin.Context) {
	userID := c.GetUint("user_id")
	apptID, _ := strconv.Atoi(c.Param("id"))

	var appointment models.Appointment
	if err := models.DB.Where("id = ? AND user_id = ?", apptID, userID).First(&appointment).Error; err != nil {
		utils.NotFound(c, "预约记录不存在")
		return
	}

	if appointment.Status != "pending" && appointment.Status != "confirmed" {
		utils.BadRequest(c, "当前状态无法取消")
		return
	}

	var req struct {
		CancelReason string `json:"cancel_reason"`
	}
	c.ShouldBindJSON(&req)

	appointment.Status = "cancelled"
	appointment.CancelReason = req.CancelReason

	if err := models.DB.Save(&appointment).Error; err != nil {
		utils.InternalError(c, "取消预约失败: "+err.Error())
		return
	}

	utils.UpdateAppointmentReminder(&appointment)

	utils.SuccessWithMessage(c, "取消成功", appointment)
}

func RescheduleAppointment(c *gin.Context) {
	userID := c.GetUint("user_id")
	apptID, _ := strconv.Atoi(c.Param("id"))

	var appointment models.Appointment
	if err := models.DB.Where("id = ? AND user_id = ?", apptID, userID).First(&appointment).Error; err != nil {
		utils.NotFound(c, "预约记录不存在")
		return
	}

	if appointment.Status != "pending" && appointment.Status != "confirmed" {
		utils.BadRequest(c, "当前状态无法改签")
		return
	}

	var req struct {
		NewDate string `json:"new_date" binding:"required"`
		NewShift string `json:"new_shift" binding:"required"`
		NewDoctorID uint `json:"new_doctor_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	t, err := time.Parse("2006-01-02", req.NewDate)
	if err != nil {
		utils.BadRequest(c, "日期格式错误，请使用 YYYY-MM-DD 格式")
		return
	}

	weekdayMap := map[int]string{
		0: "周日", 1: "周一", 2: "周二", 3: "周三", 4: "周四", 5: "周五", 6: "周六",
	}
	weekday := weekdayMap[int(t.Weekday())]

	doctorID := appointment.DoctorID
	if req.NewDoctorID > 0 {
		doctorID = req.NewDoctorID
	}

	var schedule models.Schedule
	if err := models.DB.Where("doctor_id = ? AND weekday = ? AND shift = ?", doctorID, weekday, req.NewShift).First(&schedule).Error; err != nil {
		utils.BadRequest(c, "该医生该时段无排班")
		return
	}

	var bookedCount int64
	models.DB.Model(&models.Appointment{}).Where(
		"doctor_id = ? AND appointment_date = ? AND shift = ? AND status IN ? AND id != ?",
		doctorID, req.NewDate, req.NewShift, []string{"pending", "confirmed"}, apptID,
	).Count(&bookedCount)

	if bookedCount >= int64(schedule.MaxSlots) {
		utils.BadRequest(c, "该时段已约满")
		return
	}

	appointment.AppointmentDate = req.NewDate
	appointment.Shift = req.NewShift
	appointment.DoctorID = doctorID
	appointment.SlotNumber = int(bookedCount) + 1
	appointment.Remark = appointment.Remark + fmt.Sprintf(" | 改签自原预约")

	if err := models.DB.Save(&appointment).Error; err != nil {
		utils.InternalError(c, "改签失败: "+err.Error())
		return
	}

	utils.UpdateAppointmentReminder(&appointment)

	utils.SuccessWithMessage(c, "改签成功", appointment)
}

func UpdateAppointmentStatus(c *gin.Context) {
	apptID, _ := strconv.Atoi(c.Param("id"))

	var appointment models.Appointment
	if err := models.DB.First(&appointment, apptID).Error; err != nil {
		utils.NotFound(c, "预约记录不存在")
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
		"pending":   true,
		"confirmed": true,
		"completed": true,
		"cancelled": true,
		"no_show":   true,
	}

	if !validStatuses[req.Status] {
		utils.BadRequest(c, "无效的状态")
		return
	}

	appointment.Status = req.Status

	if err := models.DB.Save(&appointment).Error; err != nil {
		utils.InternalError(c, "更新状态失败: "+err.Error())
		return
	}

	if req.Status == "cancelled" || req.Status == "completed" || req.Status == "no_show" {
		utils.UpdateAppointmentReminder(&appointment)
	}

	utils.Success(c, appointment)
}
