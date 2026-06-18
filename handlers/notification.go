package handlers

import (
	"pet-hospital/models"
	"pet-hospital/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetNotifications(c *gin.Context) {
	userID := c.GetUint("user_id")
	isRead := c.Query("is_read")
	isSent := c.Query("is_sent")
	notifType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var notifications []models.Notification
	var total int64

	query := models.DB.Preload("Appointment").Preload("Appointment.Pet").Preload("Appointment.Doctor").
		Where("user_id = ?", userID)

	if isRead != "" {
		readVal := isRead == "true"
		query = query.Where("is_read = ?", readVal)
	}
	if isSent != "" {
		sentVal := isSent == "true"
		query = query.Where("is_sent = ?", sentVal)
	}
	if notifType != "" {
		query = query.Where("type = ?", notifType)
	}

	query.Model(&models.Notification{}).Count(&total)
	offset := (page - 1) * pageSize

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&notifications).Error; err != nil {
		utils.InternalError(c, "获取通知列表失败: "+err.Error())
		return
	}

	var unreadCount int64
	models.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&unreadCount)

	utils.Success(c, gin.H{
		"list":         notifications,
		"total":        total,
		"unread_count": unreadCount,
		"page":         page,
		"page_size":    pageSize,
		"total_page":   (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

func GetNotificationByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	notifID, _ := strconv.Atoi(c.Param("id"))

	var notification models.Notification
	if err := models.DB.Preload("Appointment").Preload("Appointment.Pet").Preload("Appointment.Doctor").
		Where("id = ? AND user_id = ?", notifID, userID).First(&notification).Error; err != nil {
		utils.NotFound(c, "通知不存在")
		return
	}

	if !notification.IsRead {
		now := time.Now()
		models.DB.Model(&notification).Updates(map[string]interface{}{
			"is_read":   true,
			"read_time": now,
		})
		notification.IsRead = true
		notification.ReadTime = &now
	}

	utils.Success(c, notification)
}

func MarkNotificationRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	notifID, _ := strconv.Atoi(c.Param("id"))

	var notification models.Notification
	if err := models.DB.Where("id = ? AND user_id = ?", notifID, userID).First(&notification).Error; err != nil {
		utils.NotFound(c, "通知不存在")
		return
	}

	now := time.Now()
	if err := models.DB.Model(&notification).Updates(map[string]interface{}{
		"is_read":   true,
		"read_time": now,
	}).Error; err != nil {
		utils.InternalError(c, "标记已读失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "标记成功", nil)
}

func MarkAllNotificationsRead(c *gin.Context) {
	userID := c.GetUint("user_id")

	now := time.Now()
	if err := models.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read":   true,
			"read_time": now,
		}).Error; err != nil {
		utils.InternalError(c, "批量标记已读失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "已全部标记为已读", nil)
}

func GetUnreadCount(c *gin.Context) {
	userID := c.GetUint("user_id")

	var unreadCount int64
	models.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&unreadCount)

	var unsentCount int64
	models.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_sent = ? AND is_read = ?", userID, false, false).
		Count(&unsentCount)

	utils.Success(c, gin.H{
		"unread_count": unreadCount,
		"unsent_count": unsentCount,
	})
}

func DeleteNotification(c *gin.Context) {
	userID := c.GetUint("user_id")
	notifID, _ := strconv.Atoi(c.Param("id"))

	var notification models.Notification
	if err := models.DB.Where("id = ? AND user_id = ?", notifID, userID).First(&notification).Error; err != nil {
		utils.NotFound(c, "通知不存在")
		return
	}

	if err := models.DB.Delete(&notification).Error; err != nil {
		utils.InternalError(c, "删除通知失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

func ClearAllNotifications(c *gin.Context) {
	userID := c.GetUint("user_id")

	if err := models.DB.Where("user_id = ?", userID).
		Delete(&models.Notification{}).Error; err != nil {
		utils.InternalError(c, "清空通知失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "已清空所有通知", nil)
}

func CreateTestReminder(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		AppointmentID uint `json:"appointment_id" binding:"required"`
		MinutesBefore int  `json:"minutes_before"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var appointment models.Appointment
	if err := models.DB.Where("id = ? AND user_id = ?", req.AppointmentID, userID).First(&appointment).Error; err != nil {
		utils.NotFound(c, "预约记录不存在")
		return
	}

	var pet models.Pet
	var doctor models.Doctor
	models.DB.First(&pet, appointment.PetID)
	models.DB.First(&doctor, appointment.DoctorID)

	minutes := 5
	if req.MinutesBefore > 0 {
		minutes = req.MinutesBefore
	}
	scheduledTime := time.Now().Add(time.Duration(minutes) * time.Minute)

	shiftText := map[bool]string{true: "08:00-12:00", false: "14:00-18:00"}[appointment.Shift == "上午"]

	title := "就诊提醒 - " + pet.Name
	content := "您的宠物【" + pet.Name + "】预约了【" + doctor.Name + "】医生，就诊时间：" +
		appointment.AppointmentDate + " " + appointment.Shift + " (" + shiftText + ")，请提前15分钟到院等候。"

	notification := models.Notification{
		UserID:        userID,
		AppointmentID: appointment.ID,
		Type:          "test_reminder",
		Title:         title,
		Content:       content,
		ScheduledTime: scheduledTime,
		IsSent:        false,
		IsRead:        false,
		Channel:       "in_app",
		Remark:        "测试提醒：" + strconv.Itoa(minutes) + "分钟后发送",
	}

	if err := models.DB.Create(&notification).Error; err != nil {
		utils.InternalError(c, "创建测试提醒失败: "+err.Error())
		return
	}

	utils.Success(c, notification)
}
