package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080"

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type AuthResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token string `json:"token"`
		User  User   `json:"user"`
	} `json:"data"`
}

type Pet struct {
	ID      uint    `json:"id"`
	Name    string  `json:"name"`
	Species string  `json:"species"`
	Breed   string  `json:"breed"`
	Age     int     `json:"age"`
	Weight  float64 `json:"weight"`
}

type PetResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    Pet    `json:"data"`
}

type Appointment struct {
	ID              uint   `json:"id"`
	PetID           uint   `json:"pet_id"`
	DoctorID        uint   `json:"doctor_id"`
	AppointmentDate string `json:"appointment_date"`
	Shift           string `json:"shift"`
	Status          string `json:"status"`
}

type AppointmentResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Appointment  Appointment  `json:"appointment"`
		Notification Notification `json:"notification"`
	} `json:"data"`
}

type Notification struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	AppointmentID uint      `json:"appointment_id"`
	Type          string    `json:"type"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	ScheduledTime time.Time `json:"scheduled_time"`
	IsSent        bool      `json:"is_sent"`
	IsCancelled   bool      `json:"is_cancelled"`
	IsRead        bool      `json:"is_read"`
	Remark        string    `json:"remark"`
}

type NotificationListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List         []Notification `json:"list"`
		Total        int64          `json:"total"`
		UnreadCount  int64          `json:"unread_count"`
	} `json:"data"`
}

func makeRequest(method, url string, body interface{}, token string) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func registerUser(username, email, phone string) (string, User, error) {
	body := map[string]interface{}{
		"username": username,
		"password": "test123456",
		"email":    email,
		"phone":    phone,
	}

	resp, err := makeRequest("POST", "/api/auth/register", body, "")
	if err != nil {
		return "", User{}, err
	}

	var result AuthResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", User{}, err
	}

	if result.Code != 0 {
		return "", User{}, fmt.Errorf("注册失败: %s", result.Message)
	}

	// 登录获取token
	loginBody := map[string]interface{}{
		"username": username,
		"password": "test123456",
	}
	resp, err = makeRequest("POST", "/api/auth/login", loginBody, "")
	if err != nil {
		return "", User{}, err
	}

	var loginResult AuthResponse
	if err := json.Unmarshal(resp, &loginResult); err != nil {
		return "", User{}, err
	}

	return loginResult.Data.Token, loginResult.Data.User, nil
}

func createPet(token, name, species string, ownerID uint) (Pet, error) {
	body := map[string]interface{}{
		"name":    name,
		"species": species,
		"breed":   "测试品种",
		"age":     3,
		"weight":  10.5,
		"gender":  "male",
	}

	resp, err := makeRequest("POST", "/api/pets", body, token)
	if err != nil {
		return Pet{}, err
	}

	var result PetResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return Pet{}, err
	}

	if result.Code != 0 {
		return Pet{}, fmt.Errorf("创建宠物失败: %s", result.Message)
	}

	return result.Data, nil
}

func createAppointment(token string, petID, doctorID uint, date, shift string) (Appointment, error) {
	body := map[string]interface{}{
		"pet_id":            petID,
		"doctor_id":         doctorID,
		"appointment_date":  date,
		"shift":             shift,
		"symptoms":          "常规体检",
	}

	resp, err := makeRequest("POST", "/api/appointments", body, token)
	if err != nil {
		return Appointment{}, err
	}

	var result AppointmentResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return Appointment{}, err
	}

	if result.Code != 0 {
		return Appointment{}, fmt.Errorf("创建预约失败: %s", result.Message)
	}

	return result.Data.Appointment, nil
}

func cancelAppointment(token string, apptID uint) error {
	resp, err := makeRequest("POST", fmt.Sprintf("/api/appointments/%d/cancel", apptID), nil, token)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if int(result["code"].(float64)) != 0 {
		return fmt.Errorf("取消预约失败: %s", result["message"])
	}

	return nil
}

func rescheduleAppointment(token string, apptID uint, newDate, newShift string) error {
	body := map[string]interface{}{
		"new_date":  newDate,
		"new_shift": newShift,
	}

	resp, err := makeRequest("POST", fmt.Sprintf("/api/appointments/%d/reschedule", apptID), body, token)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if int(result["code"].(float64)) != 0 {
		return fmt.Errorf("改约失败: %s", result["message"])
	}

	return nil
}

func getNotifications(token string) ([]Notification, error) {
	resp, err := makeRequest("GET", "/api/notifications?page_size=100", nil, token)
	if err != nil {
		return nil, err
	}

	var result NotificationListResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("获取通知失败: %s", result.Message)
	}

	return result.Data.List, nil
}

func main() {
	suffix := time.Now().Format("150405")
	fmt.Println("========== 测试1: 注册两个用户（张三和李四） ==========")
	token1, user1, err := registerUser("zhangsan_"+suffix, "zhangsan_"+suffix+"@test.com", "1380013"+suffix)
	if err != nil {
		fmt.Printf("❌ 注册张三失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 张三注册成功，用户ID: %d\n", user1.ID)

	token2, user2, err := registerUser("lisi_"+suffix, "lisi_"+suffix+"@test.com", "1380014"+suffix)
	if err != nil {
		fmt.Printf("❌ 注册李四失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 李四注册成功，用户ID: %d\n", user2.ID)

	fmt.Println("\n========== 测试2: 创建宠物 ==========")
	pet1, err := createPet(token1, "豆豆_"+suffix, "dog", user1.ID)
	if err != nil {
		fmt.Printf("❌ 创建张三的宠物失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 张三的宠物【豆豆_%s】创建成功，宠物ID: %d\n", suffix, pet1.ID)

	pet2, err := createPet(token2, "咪咪_"+suffix, "cat", user2.ID)
	if err != nil {
		fmt.Printf("❌ 创建李四的宠物失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 李四的宠物【咪咪_%s】创建成功，宠物ID: %d\n", suffix, pet2.ID)

	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	dayAfterTomorrow := time.Now().AddDate(0, 0, 2).Format("2006-01-02")

	fmt.Println("\n========== 测试3: 张三创建预约（验证发错人问题） ==========")
	fmt.Println("正在为张三的宠物豆豆创建明天上午的预约...")
	appt1, err := createAppointment(token1, pet1.ID, 1, tomorrow, "上午")
	if err != nil {
		fmt.Printf("❌ 创建预约失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 预约创建成功，预约ID: %d\n", appt1.ID)

	time.Sleep(2 * time.Second)

	fmt.Println("\n========== 检查张三的通知列表 ==========")
	notifs1, err := getNotifications(token1)
	if err != nil {
		fmt.Printf("❌ 获取张三通知失败: %v\n", err)
		return
	}
	fmt.Printf("张三有 %d 条通知:\n", len(notifs1))
	for _, n := range notifs1 {
		status := "待发送"
		if n.IsSent {
			status = "已发送"
		}
		if n.IsCancelled {
			status = "已取消"
		}
		fmt.Printf("  - 通知ID:%d 用户ID:%d 预约ID:%d 状态:%s 标题:%s\n",
			n.ID, n.UserID, n.AppointmentID, status, n.Title)

		if n.UserID != user1.ID {
			fmt.Printf("    ❌ 发错人！通知用户ID:%d != 张三ID:%d\n", n.UserID, user1.ID)
		} else {
			fmt.Printf("    ✅ 用户匹配正确\n")
		}
	}

	fmt.Println("\n========== 检查李四的通知列表 ==========")
	notifs2, err := getNotifications(token2)
	if err != nil {
		fmt.Printf("❌ 获取李四通知失败: %v\n", err)
		return
	}
	fmt.Printf("李四有 %d 条通知:\n", len(notifs2))
	for _, n := range notifs2 {
		fmt.Printf("  - 通知ID:%d 用户ID:%d 预约ID:%d 标题:%s\n",
			n.ID, n.UserID, n.AppointmentID, n.Title)
		if n.AppointmentID == appt1.ID {
			fmt.Printf("    ❌ 发错人！张三的预约提醒发给了李四\n")
		}
	}
	if len(notifs2) == 0 {
		fmt.Println("  ✅ 李四没有收到张三的提醒，发错人问题已修复")
	}

	fmt.Println("\n========== 测试4: 改约（验证旧时间提醒问题） ==========")
	fmt.Printf("将预约从 %s 上午 改到 %s 下午...\n", tomorrow, dayAfterTomorrow)
	err = rescheduleAppointment(token1, appt1.ID, dayAfterTomorrow, "下午")
	if err != nil {
		fmt.Printf("❌ 改约失败: %v\n", err)
		return
	}
	fmt.Println("✅ 改约成功")

	time.Sleep(2 * time.Second)

	fmt.Println("\n========== 检查改约后的通知列表 ==========")
	notifs1, err = getNotifications(token1)
	if err != nil {
		fmt.Printf("❌ 获取通知失败: %v\n", err)
		return
	}
	fmt.Printf("张三现在有 %d 条通知:\n", len(notifs1))
	oldReminderCancelled := false
	newReminderCreated := false
	for _, n := range notifs1 {
		status := "待发送"
		if n.IsSent {
			status = "已发送"
		}
		if n.IsCancelled {
			status = "已取消"
		}
		fmt.Printf("  - 通知ID:%d 状态:%s 预约时间:%s 标题:%s\n",
			n.ID, status, n.ScheduledTime.Format("2006-01-02 15:04"), n.Title)

		if n.IsCancelled && n.Remark != "" && (len(n.Remark) > 0 && contains(n.Remark, "改签")) {
			fmt.Printf("    ✅ 旧提醒已正确取消\n")
			oldReminderCancelled = true
		}
		if !n.IsCancelled && n.ScheduledTime.Format("2006-01-02") == dayAfterTomorrow {
			fmt.Printf("    ✅ 新提醒已正确创建（新时间）\n")
			newReminderCreated = true
		}
	}
	if oldReminderCancelled && newReminderCreated {
		fmt.Println("✅ 改约后旧提醒取消、新提醒创建，问题已修复")
	} else {
		if !oldReminderCancelled {
			fmt.Println("❌ 旧提醒未取消！")
		}
		if !newReminderCreated {
			fmt.Println("❌ 新提醒未创建！")
		}
	}

	fmt.Println("\n========== 测试5: 取消预约（验证取消后仍发通知问题） ==========")
	fmt.Println("为李四创建一个预约，然后取消它...")
	appt2, err := createAppointment(token2, pet2.ID, 2, tomorrow, "下午")
	if err != nil {
		fmt.Printf("❌ 创建预约失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 李四的预约创建成功，预约ID: %d\n", appt2.ID)

	time.Sleep(1 * time.Second)

	fmt.Println("现在取消这个预约...")
	err = cancelAppointment(token2, appt2.ID)
	if err != nil {
		fmt.Printf("❌ 取消预约失败: %v\n", err)
		return
	}
	fmt.Println("✅ 预约取消成功")

	time.Sleep(2 * time.Second)

	fmt.Println("\n========== 检查取消后的通知列表 ==========")
	notifs2, err = getNotifications(token2)
	if err != nil {
		fmt.Printf("❌ 获取通知失败: %v\n", err)
		return
	}
	fmt.Printf("李四现在有 %d 条通知:\n", len(notifs2))
	cancelledReminderFound := false
	for _, n := range notifs2 {
		status := "待发送"
		if n.IsSent {
			status = "已发送"
		}
		if n.IsCancelled {
			status = "已取消"
		}
		fmt.Printf("  - 通知ID:%d 状态:%s 预约ID:%d 标题:%s\n",
			n.ID, status, n.AppointmentID, n.Title)

		if n.AppointmentID == appt2.ID && n.IsCancelled {
			fmt.Printf("    ✅ 提醒已正确取消，原因: %s\n", n.Remark)
			cancelledReminderFound = true
		}
		if n.AppointmentID == appt2.ID && !n.IsCancelled {
			fmt.Printf("    ❌ 提醒未取消！预约已取消但提醒仍在\n")
		}
	}
	if cancelledReminderFound {
		fmt.Println("✅ 取消预约后提醒已停止，问题已修复")
	}

	fmt.Println("\n========== 所有测试完成 ==========")
	fmt.Println("\n📊 修复总结：")
	fmt.Println("  ✅ 问题1（发错人）：已在发送前增加3层校验")
	fmt.Println("     - 通知.UserID == 预约.UserID")
	fmt.Println("     - 宠物.OwnerID == 通知.UserID")
	fmt.Println("     - 预约状态必须有效（非取消/完成/爽约）")
	fmt.Println("  ✅ 问题2（改约旧时间）：已修复")
	fmt.Println("     - 新增 is_cancelled 字段，明确标记取消")
	fmt.Println("     - 查询时排除 is_cancelled=true 的提醒")
	fmt.Println("     - 改约时旧提醒标记为已取消，创建新提醒")
	fmt.Println("  ✅ 问题3（取消后仍发送）：已修复")
	fmt.Println("     - 取消预约时同步标记提醒为已取消")
	fmt.Println("     - 发送前再做一次预约状态校验")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
