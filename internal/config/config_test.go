package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		fileExists    bool
		expectedError string
		expectedCfg   *Config
	}{
		{
			name: "Valid config",
			configContent: `
server:
  port: ":8080"
database:
  host: localhost
  port: 5432
  name: testdb
  user: testuser
  password: testpass
`,
			fileExists:    true,
			expectedError: "",
			expectedCfg: &Config{
				Server: ServerConfig{Port: ":8080"},
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					Name:     "testdb",
					User:     "testuser",
					Password: "testpass",
				},
			},
		},
		{
			name:          "File does not exist",
			configContent: "",
			fileExists:    false,
			expectedError: "failed to read config file",
			expectedCfg:   nil,
		},
		{
			name: "Invalid YAML",
			configContent: `
server:
  port: ":8080"
database:
  host: localhost
  port: invalid_port
  name: testdb
  user: testuser
  password: testpass
`,
			fileExists:    true,
			expectedError: "failed to unmarshal config: yaml: unmarshal errors",
			expectedCfg:   nil,
		},
		{
			name: "Missing server port",
			configContent: `
server:
  port: ""
database:
  host: localhost
  port: 5432
  name: testdb
  user: testuser
  password: testpass
`,
			fileExists:    true,
			expectedError: "server port is required",
			expectedCfg:   nil,
		},
		{
			name: "Missing database fields",
			configContent: `
server:
  port: ":8080"
database:
  host: ""
  port: 5432
  name: testdb
  user: testuser
  password: testpass
`,
			fileExists:    true,
			expectedError: "database host, name, user, and password are required",
			expectedCfg:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем временный файл для теста
			var tempFilePath string
			if tt.fileExists {
				tempFile, err := os.CreateTemp("", "config-*.yaml")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer os.Remove(tempFile.Name())

				if _, err := tempFile.WriteString(tt.configContent); err != nil {
					t.Fatalf("Failed to write to temp file: %v", err)
				}
				tempFilePath = tempFile.Name()
				tempFile.Close()
			} else {
				tempFilePath = filepath.Join(os.TempDir(), "nonexistent.yaml")
			}

			// Вызываем тестируемую функцию
			cfg, err := LoadConfig(tempFilePath)

			// Проверяем ошибку
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.expectedError)
				} else if !containsError(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tt.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Проверяем результат
			if tt.expectedCfg == nil && cfg != nil {
				t.Errorf("Expected nil config, got %v", cfg)
			} else if tt.expectedCfg != nil {
				if cfg == nil {
					t.Errorf("Expected config, got nil")
				} else {
					if cfg.Server.Port != tt.expectedCfg.Server.Port {
						t.Errorf("Expected server port %q, got %q", tt.expectedCfg.Server.Port, cfg.Server.Port)
					}
					if cfg.Database.Host != tt.expectedCfg.Database.Host ||
						cfg.Database.Port != tt.expectedCfg.Database.Port ||
						cfg.Database.Name != tt.expectedCfg.Database.Name ||
						cfg.Database.User != tt.expectedCfg.Database.User ||
						cfg.Database.Password != tt.expectedCfg.Database.Password {
						t.Errorf("Database config mismatch: expected %v, got %v", tt.expectedCfg.Database, cfg.Database)
					}
				}
			}
		})
	}
}

// Вспомогательная функция для проверки частичного совпадения ошибки
func containsError(actual, expected string) bool {
	return len(expected) <= len(actual) && actual[:len(expected)] == expected
}
