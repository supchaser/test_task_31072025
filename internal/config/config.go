package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	LogMode    string
	ServerPort string
}

func checkEnv(envVars []string) error {
	var missingVars []string

	for _, envVar := range envVars {
		if value, exists := os.LookupEnv(envVar); !exists || value == "" {
			missingVars = append(missingVars, envVar)
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("error: this env vars are missing: %v", missingVars)
	} else {
		return nil
	}
}

func validateEnv() error {
	err := checkEnv([]string{
		"LOG_MODE",
		"SERVER_PORT",
	})
	if err != nil {
		return err
	}

	return nil
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, fmt.Errorf("load cofiguration file: %w", err)
	}

	err = validateEnv()
	if err != nil {
		return nil, fmt.Errorf("LoadConfig: %w", err)
	}

	return &Config{
		LogMode:    os.Getenv("LOG_MODE"),
		ServerPort: os.Getenv("SERVER_PORT"),
	}, nil
}
