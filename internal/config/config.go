package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
}

var Cfg Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env файл не найден, используются переменные окружения")
	}

	Cfg = Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "subscriptions"),
		ServerPort: getEnv("PORT", "8081"),
	}
}

func getEnv(env, defV string) string {
	if v, exists := os.LookupEnv(env); exists {
		return v
	}
	return defV
}
