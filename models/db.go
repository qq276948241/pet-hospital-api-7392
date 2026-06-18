package models

import (
	"database/sql"
	"log"
	"pet-hospital/config"

	_ "modernc.org/sqlite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	sqlDB, err := sql.Open("sqlite", config.AppConfig.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	DB, err = gorm.Open(sqlite.Dialector{
		Conn: sqlDB,
	}, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(
		&User{},
		&Pet{},
		&Doctor{},
		&Schedule{},
		&Appointment{},
		&MedicalRecord{},
		&Prescription{},
		&PrescriptionItem{},
		&Bill{},
		&BillItem{},
		&Vaccine{},
		&VaccineRecord{},
		&Medicine{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	seedInitialData()
}

func seedInitialData() {
	var doctorCount int64
	DB.Model(&Doctor{}).Count(&doctorCount)
	if doctorCount == 0 {
		doctors := []Doctor{
			{Name: "张医生", Specialization: "全科", Phone: "13800138001", Title: "主治医师"},
			{Name: "李医生", Specialization: "外科", Phone: "13800138002", Title: "副主任医师"},
			{Name: "王医生", Specialization: "内科", Phone: "13800138003", Title: "主任医师"},
			{Name: "赵医生", Specialization: "皮肤科", Phone: "13800138004", Title: "主治医师"},
			{Name: "陈医生", Specialization: "牙科", Phone: "13800138005", Title: "主治医师"},
		}
		DB.Create(&doctors)
	}

	var medicineCount int64
	DB.Model(&Medicine{}).Count(&medicineCount)
	if medicineCount == 0 {
		medicines := []Medicine{
			{Name: "阿莫西林", Category: "抗生素", Price: 25.50, Stock: 100, Unit: "盒", Description: "广谱抗生素"},
			{Name: "头孢氨苄", Category: "抗生素", Price: 35.00, Stock: 80, Unit: "盒", Description: "头孢类抗生素"},
			{Name: "布洛芬", Category: "止痛药", Price: 15.80, Stock: 150, Unit: "盒", Description: "解热镇痛"},
			{Name: "蒙脱石散", Category: "消化系统", Price: 12.50, Stock: 120, Unit: "盒", Description: "止泻药"},
			{Name: "益生菌", Category: "消化系统", Price: 45.00, Stock: 60, Unit: "盒", Description: "调节肠道菌群"},
			{Name: "碘伏", Category: "外用消毒", Price: 8.50, Stock: 200, Unit: "瓶", Description: "皮肤消毒"},
			{Name: "眼药水", Category: "眼科用药", Price: 28.00, Stock: 50, Unit: "瓶", Description: "抗菌消炎眼药水"},
			{Name: "滴耳液", Category: "耳科用药", Price: 32.00, Stock: 45, Unit: "瓶", Description: "治疗耳部炎症"},
		}
		DB.Create(&medicines)
	}

	var vaccineCount int64
	DB.Model(&Vaccine{}).Count(&vaccineCount)
	if vaccineCount == 0 {
		vaccines := []Vaccine{
			{Name: "狂犬疫苗", Type: "核心疫苗", Price: 80.00, DurationDays: 365, Description: "预防狂犬病"},
			{Name: "四联疫苗", Type: "核心疫苗", Price: 120.00, DurationDays: 365, Description: "预防犬瘟热、细小病毒等"},
			{Name: "六联疫苗", Type: "核心疫苗", Price: 150.00, DurationDays: 365, Description: "预防多种传染病"},
			{Name: "猫三联", Type: "核心疫苗", Price: 100.00, DurationDays: 365, Description: "预防猫瘟、猫鼻支等"},
			{Name: "体内驱虫", Type: "驱虫", Price: 60.00, DurationDays: 90, Description: "体内寄生虫驱虫"},
			{Name: "体外驱虫", Type: "驱虫", Price: 80.00, DurationDays: 30, Description: "体外寄生虫驱虫"},
		}
		DB.Create(&vaccines)
	}

	var scheduleCount int64
	DB.Model(&Schedule{}).Count(&scheduleCount)
	if scheduleCount == 0 {
		weekdays := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
		shifts := []string{"上午", "下午"}

		for i := 1; i <= 5; i++ {
			for _, weekday := range weekdays {
				for _, shift := range shifts {
					if i%2 == 0 && weekday == "周日" {
						continue
					}
					if i%3 == 0 && weekday == "周六" && shift == "下午" {
						continue
					}
					schedule := Schedule{
						DoctorID: uint(i),
						Weekday:  weekday,
						Shift:    shift,
						MaxSlots: 10,
					}
					DB.Create(&schedule)
				}
			}
		}
	}
}
