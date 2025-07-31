package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	LogMode        string
	ServerPort     string
	MaxActiveTasks int
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
		"MAX_ACTIVE_TASKS",
	})
	if err != nil {
		return err
	}

	return nil
}

func LoadConfig(envFile string) (*Config, error) {
	err := godotenv.Load(envFile)
	if err != nil {
		return nil, fmt.Errorf("load configuration file: %w", err)
	}

	err = validateEnv()
	if err != nil {
		return nil, fmt.Errorf("LoadConfig: %w", err)
	}

	return &Config{
		LogMode:        os.Getenv("LOG_MODE"),
		ServerPort:     os.Getenv("SERVER_PORT"),
		MaxActiveTasks: stringToInt(os.Getenv("MAX_ACTIVE_TASKS")),
	}, nil
}

func stringToInt(s string) int {
	i, _ := strconv.ParseInt(s, 10, 32)
	return int(i)
}
