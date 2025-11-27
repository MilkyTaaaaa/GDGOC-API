package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	AppEnv      string
	DatabaseURL string
	GeminiAPIKey	string
}

var AppConfig *Config

func LoadConfig() {
	// ngeload .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	AppConfig = &Config{
		Port:        getEnv("PORT", "3000"),
		AppEnv:      getEnv("APP_ENV", "development"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
	}

	// validasi konfig
	if AppConfig.GeminiAPIKey == ""{
		log.Println(" GEMINI_API_KEY belum di set.")
	} else{
		log.Println("Gemini API berhasil di load!")
	}
	
	if AppConfig.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	log.Printf("✅ Configuration loaded successfully")
	log.Printf("   - Port: %s", AppConfig.Port)
	log.Printf("   - Environment: %s", AppConfig.AppEnv)
}

// ngambil nilai env variabel
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetConfig returns the application configuration
func GetConfig() *Config {
	return AppConfig
}