package config

import (
	"os"
	"testing"
)

func TestCheckEnv(t *testing.T) {
	tests := []struct {
		name      string
		envVars   []string
		setup     func()
		teardown  func()
		wantError bool
	}{
		{
			name:    "AllVariablesPresent",
			envVars: []string{"TEST_VAR_1", "TEST_VAR_2"},
			setup: func() {
				os.Setenv("TEST_VAR_1", "value1")
				os.Setenv("TEST_VAR_2", "value2")
			},
			teardown: func() {
				os.Unsetenv("TEST_VAR_1")
				os.Unsetenv("TEST_VAR_2")
			},
			wantError: false,
		},
		{
			name:    "OneVariableMissing",
			envVars: []string{"TEST_VAR_1", "TEST_VAR_2"},
			setup: func() {
				os.Setenv("TEST_VAR_1", "value1")
			},
			teardown: func() {
				os.Unsetenv("TEST_VAR_1")
			},
			wantError: true,
		},
		{
			name:    "VariablePresentButEmpty",
			envVars: []string{"TEST_VAR_1"},
			setup: func() {
				os.Setenv("TEST_VAR_1", "")
			},
			teardown: func() {
				os.Unsetenv("TEST_VAR_1")
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			defer func() {
				if tt.teardown != nil {
					tt.teardown()
				}
			}()

			err := checkEnv(tt.envVars)
			if (err != nil) != tt.wantError {
				t.Errorf("checkEnv() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateEnv(t *testing.T) {
	tests := []struct {
		name      string
		setup     func()
		teardown  func()
		wantError bool
	}{
		{
			name: "AllRequiredVariablesPresent",
			setup: func() {
				os.Setenv("LOG_MODE", "debug")
				os.Setenv("SERVER_PORT", "8080")
				os.Setenv("MAX_ACTIVE_TASKS", "10")
			},
			teardown: func() {
				os.Unsetenv("LOG_MODE")
				os.Unsetenv("SERVER_PORT")
				os.Unsetenv("MAX_ACTIVE_TASKS")
			},
			wantError: false,
		},
		{
			name: "MissingOneRequiredVariable",
			setup: func() {
				os.Setenv("LOG_MODE", "debug")
				os.Setenv("SERVER_PORT", "8080")
			},
			teardown: func() {
				os.Unsetenv("LOG_MODE")
				os.Unsetenv("SERVER_PORT")
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			defer func() {
				if tt.teardown != nil {
					tt.teardown()
				}
			}()

			err := validateEnv()
			if (err != nil) != tt.wantError {
				t.Errorf("validateEnv() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestStringToInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:    "ValidNumber",
			input:   "42",
			want:    42,
			wantErr: false,
		},
		{
			name:    "InvalidNumber",
			input:   "not_a_number",
			want:    0,
			wantErr: true,
		},
		{
			name:    "EmptyString",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringToInt(tt.input)
			if (got != tt.want) && !tt.wantErr {
				t.Errorf("stringToInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	const testEnvContent = `LOG_MODE=debug
					SERVER_PORT=8080
					MAX_ACTIVE_TASKS=10
					`

	envFile, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatalf("Failed to create temp .env file: %v", err)
	}
	defer os.Remove(envFile.Name())

	if _, err := envFile.WriteString(testEnvContent); err != nil {
		t.Fatalf("Failed to write to temp .env file: %v", err)
	}
	if err := envFile.Close(); err != nil {
		t.Fatalf("Failed to close temp .env file: %v", err)
	}

	tests := []struct {
		name      string
		envFile   string
		want      *Config
		wantError bool
	}{
		{
			name:    "successful config load",
			envFile: envFile.Name(),
			want: &Config{
				LogMode:        "debug",
				ServerPort:     "8080",
				MaxActiveTasks: 10,
			},
			wantError: false,
		},
		{
			name:      "missing env file",
			envFile:   "nonexistent_file",
			want:      nil,
			wantError: true,
		},
		{
			name:      "empty env file path",
			envFile:   "",
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadConfig(tt.envFile)
			if (err != nil) != tt.wantError {
				t.Errorf("LoadConfig() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if got.LogMode != tt.want.LogMode {
					t.Errorf("LoadConfig() LogMode = %v, want %v", got.LogMode, tt.want.LogMode)
				}
				if got.ServerPort != tt.want.ServerPort {
					t.Errorf("LoadConfig() ServerPort = %v, want %v", got.ServerPort, tt.want.ServerPort)
				}
				if got.MaxActiveTasks != tt.want.MaxActiveTasks {
					t.Errorf("LoadConfig() MaxActiveTasks = %v, want %v", got.MaxActiveTasks, tt.want.MaxActiveTasks)
				}
			}
		})
	}
}
