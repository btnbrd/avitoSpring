package services

import (
	"avitoSpring/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"strings"
	"testing"
	"time"
)

const testSecretKey = "test-secret-key"

func generateTestToken(role string, secretKey string, valid bool, withRole bool) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	if withRole {
		claims["role"] = role
	}
	if !valid {
		claims["exp"] = time.Now().Add(-time.Hour).Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func generateTestTokenWithMethod(role string, secretKey string, method jwt.SigningMethod) (string, error) {
	claims := jwt.MapClaims{
		"role": role,
		"exp":  time.Now().Add(time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(method, claims)
	return token.SignedString([]byte(secretKey))
}

func TestNewJWTService(t *testing.T) {
	// Устанавливаем переменную окружения для теста
	os.Setenv("JWT_SECRET_KEY", testSecretKey)
	defer os.Unsetenv("JWT_SECRET_KEY") // Очищаем после теста

	service := NewJWTService()

	if service == nil {
		t.Fatal("Expected non-nil JWTService, got nil")
	}
	if service.secretKey != testSecretKey {
		t.Errorf("Expected secretKey %q, got %q", testSecretKey, service.secretKey)
	}
}

func TestNewJWTService_EmptySecretKey(t *testing.T) {
	// Убедимся, что переменная окружения не установлена
	os.Unsetenv("JWT_SECRET_KEY")

	service := NewJWTService()

	if service == nil {
		t.Fatal("Expected non-nil JWTService, got nil")
	}
	if service.secretKey != "" {
		t.Errorf("Expected empty secretKey, got %q", service.secretKey)
	}
}

func TestJWTService_GenerateToken(t *testing.T) {
	// Устанавливаем переменную окружения для теста
	os.Setenv("JWT_SECRET_KEY", testSecretKey)
	defer os.Unsetenv("JWT_SECRET_KEY")

	service := NewJWTService()

	tests := []struct {
		name          string
		role          models.Role
		expectedError string
	}{
		{
			name:          "Valid employee token",
			role:          models.RoleEmployee,
			expectedError: "",
		},
		{
			name:          "Valid moderator token",
			role:          models.RoleModerator,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.role)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error %q, got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if token == "" {
				t.Error("Expected non-empty token, got empty")
			}

			// Проверяем, что токен валиден и содержит правильную роль
			parsedRole, err := service.ValidateToken(token)
			if err != nil {
				t.Errorf("Failed to validate generated token: %v", err)
			}
			if parsedRole != tt.role {
				t.Errorf("Expected role %v, got %v", tt.role, parsedRole)
			}
		})
	}
}

func TestJWTService_GenerateToken_NoSecretKey(t *testing.T) {
	// Убедимся, что переменная окружения не установлена
	os.Unsetenv("JWT_SECRET_KEY")

	service := NewJWTService()

	tests := []struct {
		name          string
		role          models.Role
		expectedError string
	}{
		{
			name:          "Employee token with no secret key",
			role:          models.RoleEmployee,
			expectedError: "JWT secret key is not set",
		},
		{
			name:          "Moderator token with no secret key",
			role:          models.RoleModerator,
			expectedError: "JWT secret key is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.role)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error %q, got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error %q, got %q", tt.expectedError, err.Error())
				}
				if token != "" {
					t.Errorf("Expected empty token, got %q", token)
				}
			} else {
				t.Errorf("Expected error, got nil")
			}
		})
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	// Устанавливаем переменную окружения для теста
	os.Setenv("JWT_SECRET_KEY", testSecretKey)
	defer os.Unsetenv("JWT_SECRET_KEY")

	service := NewJWTService()

	tests := []struct {
		name          string
		tokenString   string
		expectedRole  models.Role
		expectedError string
	}{
		{
			name: "Valid employee token",
			tokenString: func() string {
				token, _ := generateTestToken(string(models.RoleEmployee), testSecretKey, true, true)
				return token
			}(),
			expectedRole:  models.RoleEmployee,
			expectedError: "",
		},
		{
			name: "Valid moderator token",
			tokenString: func() string {
				token, _ := generateTestToken(string(models.RoleModerator), testSecretKey, true, true)
				return token
			}(),
			expectedRole:  models.RoleModerator,
			expectedError: "",
		},
		{
			name:          "Invalid token format",
			tokenString:   "invalid-token",
			expectedRole:  "",
			expectedError: "token is malformed: token contains an invalid number of segments",
		},
		{
			name: "Expired token",
			tokenString: func() string {
				token, _ := generateTestToken(string(models.RoleEmployee), testSecretKey, false, true)
				return token
			}(),
			expectedRole:  "",
			expectedError: "token has invalid claims: token is expired",
		},
		{
			name: "Wrong secret key",
			tokenString: func() string {
				token, _ := generateTestToken(string(models.RoleEmployee), "wrong-key", true, true)
				return token
			}(),
			expectedRole:  "",
			expectedError: "token signature is invalid: signature is invalid",
		},
		{
			name: "Invalid role",
			tokenString: func() string {
				token, _ := generateTestToken("invalid-role", testSecretKey, true, true)
				return token
			}(),
			expectedRole:  "",
			expectedError: "invalid role in token: invalid-role",
		},
		{
			name: "Empty role",
			tokenString: func() string {
				token, _ := generateTestToken("", testSecretKey, true, true)
				return token
			}(),
			expectedRole:  "",
			expectedError: "invalid role in token: ",
		},
		{
			name: "Missing role claim",
			tokenString: func() string {
				token, _ := generateTestToken("", testSecretKey, true, false)
				return token
			}(),
			expectedRole:  "",
			expectedError: "role not found in token",
		},
		{
			name: "Wrong signing method",
			tokenString: func() string {
				token, _ := generateTestTokenWithMethod(string(models.RoleEmployee), testSecretKey, jwt.SigningMethodHS512)
				return token
			}(),
			expectedRole:  "",
			expectedError: "unexpected signing method: HS512",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := service.ValidateToken(tt.tokenString)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				if role != tt.expectedRole {
					t.Errorf("Expected role %v, got %v", tt.expectedRole, role)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if role != tt.expectedRole {
					t.Errorf("Expected role %v, got %v", tt.expectedRole, role)
				}
			}
		})
	}
}

func TestJWTService_ValidateToken_NoSecretKey(t *testing.T) {
	// Убедимся, что переменная окружения не установлена
	os.Unsetenv("JWT_SECRET_KEY")

	service := NewJWTService()

	tests := []struct {
		name          string
		tokenString   string
		expectedRole  models.Role
		expectedError string
	}{
		{
			name: "Valid token with no secret key",
			tokenString: func() string {
				token, _ := generateTestToken(string(models.RoleEmployee), testSecretKey, true, true)
				return token
			}(),
			expectedRole:  "",
			expectedError: "JWT secret key is not set",
		},
		{
			name:          "Invalid token with no secret key",
			tokenString:   "invalid-token",
			expectedRole:  "",
			expectedError: "JWT secret key is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := service.ValidateToken(tt.tokenString)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error %q, got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error %q, got %q", tt.expectedError, err.Error())
				}
				if role != tt.expectedRole {
					t.Errorf("Expected role %v, got %v", tt.expectedRole, role)
				}
			} else {
				t.Errorf("Expected error, got nil")
			}
		})
	}
}
