package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort         string
	DBUrl              string
	JwtSecret          string
	JwtExpirationHours int
	ResumeParserApiKey string
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	expHours, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "72"))
	if err != nil {
		log.Fatalf("Invalid JWT_EXPIRATION_HOURS: %v", err)
	}

	AppConfig = &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		DBUrl:              getEnv("DB_URL", ""),
		JwtSecret:          getEnv("JWT_SECRET", "default_secret"),
		JwtExpirationHours: expHours,
		ResumeParserApiKey: getEnv("RESUME_PARSER_API_KEY", ""),
	}

	if AppConfig.DBUrl == "" || AppConfig.ResumeParserApiKey == "" {
		log.Fatal("Database URL and Resume Parser API Key must be set in .env file")
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
