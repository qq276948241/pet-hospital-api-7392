package main

import (
	"log"
	"pet-hospital/config"
	"pet-hospital/models"
	"pet-hospital/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	models.InitDB()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	routes.SetupRoutes(r)

	log.Printf("Server starting on port %s...", config.AppConfig.Port)
	log.Printf("API Documentation:")
	log.Printf("  Health Check:   GET /health")
	log.Printf("  Auth:           POST /api/auth/register, /api/auth/login")
	log.Printf("  Pets:           GET/POST /api/pets")
	log.Printf("  Doctors:        GET /api/doctors")
	log.Printf("  Schedules:      GET /api/schedules")
	log.Printf("  Appointments:   GET/POST /api/appointments")
	log.Printf("  Medical Records:GET/POST /api/medical-records")
	log.Printf("  Prescriptions:  GET/POST /api/prescriptions")
	log.Printf("  Bills:          GET/POST /api/bills")
	log.Printf("  Vaccines:       GET /api/vaccines")
	log.Printf("  Medicines:      GET/POST /api/medicines")
	log.Printf("  Dashboard:      GET /api/dashboard/*")

	if err := r.Run(":" + config.AppConfig.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
