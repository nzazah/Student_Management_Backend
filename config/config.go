package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  .env file not found, using environment variables")
	}
	os.Setenv("TZ", "Asia/Jakarta")
}
