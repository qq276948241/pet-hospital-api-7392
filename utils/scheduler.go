package utils

import (
	"fmt"
	"log"
	"pet-hospital/models"
	"sync"
	"time"
)

type Scheduler struct {
	ticker  *time.Ticker
	stopCh  chan struct{}
	wg      sync.WaitGroup
	running bool
	mu      sync.Mutex
}

var globalScheduler *Scheduler
var schedulerOnce sync.Once

func GetScheduler() *Scheduler {
	schedulerOnce.Do(func() {
		globalScheduler = &Scheduler{
			stopCh: make(chan struct{}),
		}
	})
	return globalScheduler
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.ticker = time.NewTicker(30 * time.Second)
	s.mu.Unlock()

	log.Println("[Scheduler] 预约提醒调度器已启动，每30秒检查一次待发送的提醒...")

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ticker.C:
				s.checkAndSendNotifications()
			case <-s.stopCh:
				log.Println("[Scheduler] 调度器已停止")
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	s.running = false
	s.ticker.Stop()
	close(s.stopCh)
	s.wg.Wait()
}

func (s *Scheduler) checkAndSendNotifications() {
	now := time.Now()

	var notifications []models.Notification
	err := models.DB.Preload("Appointment").Preload("Appointment.Pet").Preload("Appointment.Doctor").
		Where("is_sent = ? AND is_cancelled = ? AND scheduled_time <= ?", false, false, now).
		Find(&notifications).Error
	if err != nil {
		log.Printf("[Scheduler] 查询待发送提醒失败: %v", err)
		return
	}

	if len(notifications) == 0 {
		return
	}

	log.Printf("[Scheduler] 检查到 %d 条待发送提醒", len(notifications))

	sentCount := 0
	skippedCount := 0

	for i := range notifications {
		notif := &notifications[i]

		if !validateNotificationBeforeSend(notif) {
			skippedCount++
			continue
		}

		if err := sendNotification(notif); err != nil {
			log.Printf("[Scheduler] 发送提醒失败 (ID:%d): %v", notif.ID, err)
			continue
		}

		sentTime := time.Now()
		models.DB.Model(notif).Updates(map[string]interface{}{
			"is_sent":  true,
			"sent_time": sentTime,
		})
		sentCount++
	}

	log.Printf("[Scheduler] 本次发送完成：成功 %d 条，跳过 %d 条", sentCount, skippedCount)
}

func validateNotificationBeforeSend(notif *models.Notification) bool {
	if notif.IsCancelled {
		log.Printf("[Scheduler] 跳过提醒 ID:%d：已被取消", notif.ID)
		return false
	}

	if notif.AppointmentID == 0 {
		log.Printf("[Scheduler] 跳过提醒 ID:%d：无关联预约", notif.ID)
		cancelNotification(notif, "无关联预约")
		return false
	}

	if notif.Appointment == nil {
		var appt models.Appointment
		if err := models.DB.First(&appt, notif.AppointmentID).Error; err != nil {
			log.Printf("[Scheduler] 跳过提醒 ID:%d：关联预约不存在", notif.ID)
			cancelNotification(notif, "关联预约已删除")
			return false
		}
		notif.Appointment = &appt
	}

	if notif.Appointment.UserID != notif.UserID {
		log.Printf("[Scheduler] ⚠️  发错人风险！提醒 ID:%d 所属用户:%d，但预约用户:%d，已自动取消",
			notif.ID, notif.UserID, notif.Appointment.UserID)
		cancelNotification(notif, fmt.Sprintf("用户不匹配: 通知用户%d != 预约用户%d", notif.UserID, notif.Appointment.UserID))
		return false
	}

	invalidStatuses := []string{"cancelled", "completed", "no_show"}
	for _, status := range invalidStatuses {
		if notif.Appointment.Status == status {
			log.Printf("[Scheduler] 跳过提醒 ID:%d：预约已%s", notif.ID, status)
			cancelNotification(notif, fmt.Sprintf("预约已%s", status))
			return false
		}
	}

	if notif.Appointment.PetID != 0 {
		var pet models.Pet
		if err := models.DB.First(&pet, notif.Appointment.PetID).Error; err == nil {
			if pet.OwnerID != notif.UserID {
				log.Printf("[Scheduler] ⚠️  宠物主人不匹配！提醒 ID:%d 用户:%d，但宠物主人:%d",
					notif.ID, notif.UserID, pet.OwnerID)
				cancelNotification(notif, fmt.Sprintf("宠物主人不匹配: 用户%d != 主人%d", notif.UserID, pet.OwnerID))
				return false
			}
		}
	}

	return true
}

func cancelNotification(notif *models.Notification, reason string) {
	remark := notif.Remark
	if remark != "" {
		remark += " | "
	}
	remark += "自动取消: " + reason

	models.DB.Model(notif).Updates(map[string]interface{}{
		"is_cancelled": true,
		"remark":       remark,
	})
}

func sendNotification(notif *models.Notification) error {
	switch notif.Channel {
	case "in_app", "":
		return sendInAppNotification(notif)
	case "sms":
		return sendSMSNotification(notif)
	case "email":
		return sendEmailNotification(notif)
	default:
		return sendInAppNotification(notif)
	}
}

func sendInAppNotification(notif *models.Notification) error {
	log.Printf("[通知] 发送站内信给用户ID:%d | 标题:%s | 内容:%s",
		notif.UserID, notif.Title, notif.Content)
	return nil
}

func sendSMSNotification(notif *models.Notification) error {
	log.Printf("[通知] 发送短信给用户ID:%d | 标题:%s | 内容:%s",
		notif.UserID, notif.Title, notif.Content)
	return nil
}

func sendEmailNotification(notif *models.Notification) error {
	log.Printf("[通知] 发送邮件给用户ID:%d | 标题:%s | 内容:%s",
		notif.UserID, notif.Title, notif.Content)
	return nil
}

func CreateAppointmentReminder(appointment *models.Appointment) (*models.Notification, error) {
	var pet models.Pet
	var doctor models.Doctor

	if err := models.DB.First(&pet, appointment.PetID).Error; err != nil {
		return nil, fmt.Errorf("获取宠物信息失败: %v", err)
	}
	if err := models.DB.First(&doctor, appointment.DoctorID).Error; err != nil {
		return nil, fmt.Errorf("获取医生信息失败: %v", err)
	}

	if pet.OwnerID != appointment.UserID {
		return nil, fmt.Errorf("宠物主人不匹配: 宠物主人ID=%d, 预约用户ID=%d", pet.OwnerID, appointment.UserID)
	}

	appointmentTime, err := parseAppointmentTime(appointment.AppointmentDate, appointment.Shift)
	if err != nil {
		return nil, fmt.Errorf("解析预约时间失败: %v", err)
	}

	scheduledTime := appointmentTime.Add(-1 * time.Hour)
	if scheduledTime.Before(time.Now()) {
		scheduledTime = time.Now().Add(5 * time.Minute)
	}

	shiftText := map[bool]string{true: "08:00-12:00", false: "14:00-18:00"}[appointment.Shift == "上午"]

	title := "就诊提醒 - " + pet.Name
	content := fmt.Sprintf("您的宠物【%s】预约了【%s】医生，就诊时间：%s %s (%s)，请提前15分钟到院等候。",
		pet.Name, doctor.Name, appointment.AppointmentDate, appointment.Shift, shiftText)

	notification := models.Notification{
		UserID:        appointment.UserID,
		AppointmentID: appointment.ID,
		Type:          "appointment_reminder",
		Title:         title,
		Content:       content,
		ScheduledTime: scheduledTime,
		IsSent:        false,
		IsCancelled:   false,
		IsRead:        false,
		Channel:       "in_app",
		Remark:        fmt.Sprintf("预约时间前1小时提醒 | 原预约时间: %s", appointmentTime.Format("2006-01-02 15:04")),
	}

	if err := models.DB.Create(&notification).Error; err != nil {
		return nil, fmt.Errorf("创建提醒失败: %v", err)
	}

	log.Printf("[提醒调度] 已为预约ID:%d 创建提醒 (通知ID:%d)，将在 %s 发送给用户ID:%d",
		appointment.ID, notification.ID, scheduledTime.Format("2006-01-02 15:04:05"), appointment.UserID)

	return &notification, nil
}

func UpdateAppointmentReminder(appointment *models.Appointment) error {
	var existingNotifs []models.Notification
	err := models.DB.Where("appointment_id = ? AND is_sent = ? AND is_cancelled = ? AND type = ?",
		appointment.ID, false, false, "appointment_reminder").Find(&existingNotifs).Error
	if err != nil {
		return fmt.Errorf("查询现有提醒失败: %v", err)
	}

	for i := range existingNotifs {
		notif := &existingNotifs[i]
		remark := notif.Remark
		if remark != "" {
			remark += " | "
		}

		if appointment.Status == "cancelled" {
			remark += "预约已取消"
		} else {
			remark += "预约已改签"
		}

		models.DB.Model(notif).Updates(map[string]interface{}{
			"is_cancelled": true,
			"remark":       remark,
		})
		log.Printf("[提醒调度] 已取消提醒 ID:%d，原因: %s", notif.ID, remark)
	}

	if appointment.Status == "cancelled" {
		log.Printf("[提醒调度] 预约ID:%d 已取消，已取消 %d 条关联提醒", appointment.ID, len(existingNotifs))
		return nil
	}

	newNotif, err := CreateAppointmentReminder(appointment)
	if err != nil {
		return fmt.Errorf("创建新提醒失败: %v", err)
	}

	log.Printf("[提醒调度] 预约ID:%d 已改签，已取消 %d 条旧提醒，创建新提醒 ID:%d",
		appointment.ID, len(existingNotifs), newNotif.ID)
	return nil
}

func parseAppointmentTime(dateStr, shift string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, err
	}

	loc := time.Local
	if shift == "上午" {
		return time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, loc), nil
	}
	return time.Date(date.Year(), date.Month(), date.Day(), 15, 0, 0, 0, loc), nil
}
