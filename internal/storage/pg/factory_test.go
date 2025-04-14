package pg

import (
	"avitoSpring/internal/config"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/lib/pq"
)

func TestNewStorage(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *config.Config
		mockDB        func() (*sql.DB, error) // Имитация sql.Open и db.Ping
		expectedError string
	}{

		{
			name: "Failed to ping database",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					Name:     "testdb",
					User:     "testuser",
					Password: "testpass",
				},
			},
			mockDB: func() (*sql.DB, error) {
				return nil, errors.New("failed to ping database")
			},
			expectedError: "failed to ping database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := NewStorage(tt.cfg)
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error %q, got nil", tt.expectedError)
				} else if !containsError(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tt.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func containsError(actual, expected string) bool {
	return len(expected) <= len(actual) && actual[:len(expected)] == expected
}
