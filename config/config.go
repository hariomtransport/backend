package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresURL string
	MongoURL    string
	DBType      string
	Port        string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := &Config{
		PostgresURL: os.Getenv("POSTGRES_URL"),
		MongoURL:    os.Getenv("MONGO_URL"),
		DBType:      os.Getenv("DB_TYPE"),
		Port:        os.Getenv("PORT"),
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	return cfg
}
