package config

import (
	"log"
	"os"
)

func InitLogger() {
	log.SetOutput(os.Stdout)
}
