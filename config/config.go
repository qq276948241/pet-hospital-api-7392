package config

type Config struct {
	JWTSecret string
	DBPath    string
	Port      string
}

var AppConfig = Config{
	JWTSecret: "pet-hospital-secret-key-2024",
	DBPath:    "pet_hospital.db",
	Port:      "8080",
}
