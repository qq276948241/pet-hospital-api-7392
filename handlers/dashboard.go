package handlers

import (
	"pet-hospital/models"
	"pet-hospital/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetDashboardStats(c *gin.Context) {
	var userCount int64
	models.DB.Model(&models.User{}).Count(&userCount)

	var petCount int64
	models.DB.Model(&models.Pet{}).Count(&petCount)

	var appointmentCount int64
	models.DB.Model(&models.Appointment{}).Count(&appointmentCount)

	var todayAppointments int64
	today := time.Now().Format("2006-01-02")
	models.DB.Model(&models.Appointment{}).Where("appointment_date = ?", today).Count(&todayAppointments)

	var pendingAppointments int64
	models.DB.Model(&models.Appointment{}).Where("status = ?", "pending").Count(&pendingAppointments)

	var completedAppointments int64
	models.DB.Model(&models.Appointment{}).Where("status = ?", "completed").Count(&completedAppointments)

	var totalRevenue float64
	models.DB.Model(&models.Bill{}).Where("status = ?", "paid").Select("COALESCE(SUM(total_amount), 0)").Scan(&totalRevenue)

	var unpaidBills int64
	models.DB.Model(&models.Bill{}).Where("status = ?", "unpaid").Count(&unpaidBills)

	var medicineCount int64
	models.DB.Model(&models.Medicine{}).Count(&medicineCount)

	var lowStockCount int64
	models.DB.Model(&models.Medicine{}).Where("stock < ?", 20).Count(&lowStockCount)

	utils.Success(c, gin.H{
		"total_users":          userCount,
		"total_pets":           petCount,
		"total_appointments":   appointmentCount,
		"today_appointments":   todayAppointments,
		"pending_appointments": pendingAppointments,
		"completed_appointments": completedAppointments,
		"total_revenue":        totalRevenue,
		"unpaid_bills":         unpaidBills,
		"total_medicines":      medicineCount,
		"low_stock_count":      lowStockCount,
	})
}

func GetAppointmentTrend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	type DailyCount struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	var results []DailyCount

	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i).Format("2006-01-02")
		var count int64
		models.DB.Model(&models.Appointment{}).Where("appointment_date = ?", date).Count(&count)
		results = append(results, DailyCount{Date: date, Count: count})
	}

	utils.Success(c, gin.H{
		"days": days,
		"data": results,
	})
}

func GetRevenueTrend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	type DailyRevenue struct {
		Date    string  `json:"date"`
		Revenue float64 `json:"revenue"`
		Count   int64   `json:"count"`
	}

	var results []DailyRevenue

	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i).Format("2006-01-02")
		var revenue float64
		var count int64

		models.DB.Model(&models.Bill{}).
			Where("DATE(created_at) = ? AND status = ?", date, "paid").
			Select("COALESCE(SUM(total_amount), 0)").Scan(&revenue)
		models.DB.Model(&models.Bill{}).
			Where("DATE(created_at) = ? AND status = ?", date, "paid").
			Count(&count)

		results = append(results, DailyRevenue{Date: date, Revenue: revenue, Count: count})
	}

	utils.Success(c, gin.H{
		"days": days,
		"data": results,
	})
}

func GetDoctorStats(c *gin.Context) {
	type DoctorStat struct {
		ID            uint   `json:"id"`
		Name          string `json:"name"`
		Specialization string `json:"specialization"`
		Title         string `json:"title"`
		TotalAppointments int64 `json:"total_appointments"`
		TodayAppointments int64 `json:"today_appointments"`
	}

	var doctors []models.Doctor
	models.DB.Find(&doctors)

	today := time.Now().Format("2006-01-02")
	var stats []DoctorStat

	for _, doctor := range doctors {
		var total int64
		var todayCount int64

		models.DB.Model(&models.Appointment{}).Where("doctor_id = ?", doctor.ID).Count(&total)
		models.DB.Model(&models.Appointment{}).Where("doctor_id = ? AND appointment_date = ?", doctor.ID, today).Count(&todayCount)

		stats = append(stats, DoctorStat{
			ID:                doctor.ID,
			Name:              doctor.Name,
			Specialization:    doctor.Specialization,
			Title:             doctor.Title,
			TotalAppointments: total,
			TodayAppointments: todayCount,
		})
	}

	utils.Success(c, stats)
}

func GetPetSpeciesStats(c *gin.Context) {
	type SpeciesStat struct {
		Species string `json:"species"`
		Count   int64  `json:"count"`
	}

	var stats []SpeciesStat
	models.DB.Model(&models.Pet{}).Select("species, COUNT(*) as count").Group("species").Scan(&stats)

	utils.Success(c, stats)
}

func GetMedicineUsageStats(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	type MedicineUsage struct {
		MedicineID   uint   `json:"medicine_id"`
		MedicineName string `json:"medicine_name"`
		TotalUsed    int    `json:"total_used"`
		TotalRevenue float64 `json:"total_revenue"`
	}

	var stats []MedicineUsage

	models.DB.Table("prescription_items").
		Select("medicine_id, medicine_name, SUM(quantity) as total_used, SUM(quantity * unit_price) as total_revenue").
		Group("medicine_id, medicine_name").
		Order("total_used DESC").
		Limit(limit).
		Scan(&stats)

	utils.Success(c, stats)
}

func GetAppointmentStatusStats(c *gin.Context) {
	type StatusStat struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	var stats []StatusStat
	models.DB.Model(&models.Appointment{}).Select("status, COUNT(*) as count").Group("status").Scan(&stats)

	statusMap := map[string]string{
		"pending":   "待确认",
		"confirmed": "已确认",
		"completed": "已完成",
		"cancelled": "已取消",
		"no_show":   "未到诊",
	}

	result := make([]gin.H, 0)
	for _, stat := range stats {
		label := stat.Status
		if m, ok := statusMap[stat.Status]; ok {
			label = m
		}
		result = append(result, gin.H{
			"status": stat.Status,
			"label":  label,
			"count":  stat.Count,
		})
	}

	utils.Success(c, result)
}

func GetRecentActivities(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	type Activity struct {
		ID        uint      `json:"id"`
		Type      string    `json:"type"`
		Content   string    `json:"content"`
		CreatedAt time.Time `json:"created_at"`
	}

	var activities []Activity

	var appointments []models.Appointment
	models.DB.Preload("Pet").Preload("Doctor").
		Order("created_at DESC").
		Limit(limit / 3).
		Find(&appointments)

	for _, apt := range appointments {
		activities = append(activities, Activity{
			ID:        apt.ID,
			Type:      "appointment",
			Content:   apt.Pet.Name + " 预约了 " + apt.Doctor.Name + " " + apt.AppointmentDate + " " + apt.Shift,
			CreatedAt: apt.CreatedAt,
		})
	}

	var records []models.MedicalRecord
	models.DB.Preload("Pet").Preload("Doctor").
		Order("created_at DESC").
		Limit(limit / 3).
		Find(&records)

	for _, rec := range records {
		activities = append(activities, Activity{
			ID:        rec.ID,
			Type:      "medical_record",
			Content:   rec.Pet.Name + " 的就诊记录已由 " + rec.Doctor.Name + " 创建",
			CreatedAt: rec.CreatedAt,
		})
	}

	var bills []models.Bill
	models.DB.Preload("Pet").
		Where("status = ?", "paid").
		Order("payment_time DESC").
		Limit(limit / 3).
		Find(&bills)

	for _, bill := range bills {
		activities = append(activities, Activity{
			ID:        bill.ID,
			Type:      "payment",
			Content:   bill.Pet.Name + " 的账单已支付 ¥" + strconv.FormatFloat(bill.TotalAmount, 'f', 2, 64),
			CreatedAt: *bill.PaymentTime,
		})
	}

	utils.Success(c, activities)
}

func GetUserDashboardStats(c *gin.Context) {
	userID := c.GetUint("user_id")

	var petCount int64
	models.DB.Model(&models.Pet{}).Where("owner_id = ?", userID).Count(&petCount)

	var appointmentCount int64
	models.DB.Model(&models.Appointment{}).Where("user_id = ?", userID).Count(&appointmentCount)

	var pendingAppointments int64
	models.DB.Model(&models.Appointment{}).Where("user_id = ? AND status IN ?", userID, []string{"pending", "confirmed"}).Count(&pendingAppointments)

	var unpaidAmount float64
	models.DB.Model(&models.Bill{}).Where("user_id = ? AND status = ?", userID, "unpaid").Select("COALESCE(SUM(total_amount), 0)").Scan(&unpaidAmount)

	var totalPaid float64
	models.DB.Model(&models.Bill{}).Where("user_id = ? AND status = ?", userID, "paid").Select("COALESCE(SUM(total_amount), 0)").Scan(&totalPaid)

	var upcomingAppointments []models.Appointment
	today := time.Now().Format("2006-01-02")
	models.DB.Preload("Pet").Preload("Doctor").
		Where("user_id = ? AND appointment_date >= ? AND status IN ?", userID, today, []string{"pending", "confirmed"}).
		Order("appointment_date ASC").
		Limit(5).
		Find(&upcomingAppointments)

	utils.Success(c, gin.H{
		"total_pets":            petCount,
		"total_appointments":    appointmentCount,
		"pending_appointments":  pendingAppointments,
		"unpaid_amount":         unpaidAmount,
		"total_paid":            totalPaid,
		"upcoming_appointments": upcomingAppointments,
	})
}
