package routes

import (
	"pet-hospital/handlers"
	"pet-hospital/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"status":  "running",
				"version": "1.0.0",
			},
		})
	})

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.GET("/profile", middleware.AuthMiddleware(), handlers.GetProfile)
			auth.PUT("/profile", middleware.AuthMiddleware(), handlers.UpdateProfile)
			auth.POST("/logout", middleware.AuthMiddleware(), handlers.Logout)
		}

		pets := api.Group("/pets")
		pets.Use(middleware.AuthMiddleware())
		{
			pets.POST("", handlers.CreatePet)
			pets.GET("", handlers.GetPets)
			pets.GET("/:id", handlers.GetPetByID)
			pets.PUT("/:id", handlers.UpdatePet)
			pets.DELETE("/:id", handlers.DeletePet)
		}

		doctors := api.Group("/doctors")
		{
			doctors.GET("", handlers.GetDoctors)
			doctors.GET("/:id", handlers.GetDoctorByID)
			doctors.GET("/:id/schedules", handlers.GetScheduleByDoctor)
		}

		schedules := api.Group("/schedules")
		{
			schedules.GET("", handlers.GetSchedules)
		}

		appointments := api.Group("/appointments")
		appointments.Use(middleware.AuthMiddleware())
		{
			appointments.POST("", handlers.CreateAppointment)
			appointments.GET("", handlers.GetAppointments)
			appointments.GET("/:id", handlers.GetAppointmentByID)
			appointments.POST("/:id/cancel", handlers.CancelAppointment)
			appointments.POST("/:id/reschedule", handlers.RescheduleAppointment)
			appointments.PUT("/:id/status", handlers.UpdateAppointmentStatus)
		}

		medicalRecords := api.Group("/medical-records")
		medicalRecords.Use(middleware.AuthMiddleware())
		{
			medicalRecords.POST("", handlers.CreateMedicalRecord)
			medicalRecords.GET("", handlers.GetMedicalRecords)
			medicalRecords.GET("/:id", handlers.GetMedicalRecordByID)
			medicalRecords.GET("/pet/:pet_id", handlers.GetMedicalRecordsByPet)
			medicalRecords.PUT("/:id", handlers.UpdateMedicalRecord)
		}

		prescriptions := api.Group("/prescriptions")
		prescriptions.Use(middleware.AuthMiddleware())
		{
			prescriptions.POST("", handlers.CreatePrescription)
			prescriptions.GET("", handlers.GetPrescriptions)
			prescriptions.GET("/:id", handlers.GetPrescriptionByID)
			prescriptions.GET("/pet/:pet_id", handlers.GetPrescriptionsByPet)
			prescriptions.PUT("/:id/status", handlers.UpdatePrescriptionStatus)
		}

		bills := api.Group("/bills")
		bills.Use(middleware.AuthMiddleware())
		{
			bills.POST("", handlers.CreateBill)
			bills.GET("", handlers.GetBills)
			bills.GET("/:id", handlers.GetBillByID)
			bills.POST("/:id/pay", handlers.PayBill)
			bills.GET("/pet/:pet_id", handlers.GetBillsByPet)
			bills.PUT("/:id/status", handlers.UpdateBillStatus)
		}

		vaccines := api.Group("/vaccines")
		vaccines.Use(middleware.AuthMiddleware())
		{
			vaccines.GET("", handlers.GetVaccines)
			vaccines.GET("/:id", handlers.GetVaccineByID)
			vaccines.POST("/records", handlers.AddVaccineRecord)
			vaccines.GET("/records", handlers.GetVaccineRecords)
			vaccines.GET("/records/pet/:pet_id", handlers.GetVaccineRecordsByPet)
			vaccines.GET("/reminders", handlers.GetVaccineReminders)
			vaccines.GET("/reminders/upcoming", handlers.GetUpcomingReminders)
		}

		medicines := api.Group("/medicines")
		{
			medicines.GET("", handlers.GetMedicines)
			medicines.GET("/categories", handlers.GetMedicineCategories)
			medicines.GET("/low-stock-alert", handlers.GetLowStockAlert)
			medicines.GET("/:id", handlers.GetMedicineByID)
			medicines.POST("", handlers.CreateMedicine)
			medicines.PUT("/:id", handlers.UpdateMedicine)
			medicines.PUT("/:id/stock", handlers.UpdateStock)
			medicines.POST("/:id/adjust-stock", handlers.AdjustStock)
			medicines.DELETE("/:id", handlers.DeleteMedicine)
		}

		dashboard := api.Group("/dashboard")
		dashboard.Use(middleware.AuthMiddleware())
		{
			dashboard.GET("/stats", handlers.GetDashboardStats)
			dashboard.GET("/appointment-trend", handlers.GetAppointmentTrend)
			dashboard.GET("/revenue-trend", handlers.GetRevenueTrend)
			dashboard.GET("/doctor-stats", handlers.GetDoctorStats)
			dashboard.GET("/pet-species-stats", handlers.GetPetSpeciesStats)
			dashboard.GET("/medicine-usage-stats", handlers.GetMedicineUsageStats)
			dashboard.GET("/appointment-status-stats", handlers.GetAppointmentStatusStats)
			dashboard.GET("/recent-activities", handlers.GetRecentActivities)
			dashboard.GET("/user-stats", handlers.GetUserDashboardStats)
		}
	}
}
