package services

import (
	"avitoSpring/internal/models"
	"errors"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"
)

// Мок для UserStorage
type mockUserStorage struct {
	mock.Mock
}

func (m *mockUserStorage) CreateUser(user *models.User, password string) (string, error) {
	args := m.Called(user, password)
	return args.String(0), args.Error(1)
}

func (m *mockUserStorage) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserStorage) CheckPassword(userID, password string) bool {
	args := m.Called(userID, password)
	return args.Bool(0)
}

// Мок для JWTServiceInterface
type mockJWTService struct {
	mock.Mock
}

func (m *mockJWTService) GenerateToken(role models.Role) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func (m *mockJWTService) ValidateToken(tokenString string) (models.Role, error) {
	args := m.Called(tokenString)
	return args.Get(0).(models.Role), args.Error(1)
}

func TestNewAuthService(t *testing.T) {
	store := &mockUserStorage{}
	jwtService := &mockJWTService{}
	service := NewAuthService(store, jwtService)

	if service == nil {
		t.Fatal("Expected non-nil AuthService")
	}
	if service.store != store {
		t.Errorf("Expected store %v, got %v", store, service.store)
	}
	if service.jwtService != jwtService {
		t.Errorf("Expected jwtService %v, got %v", jwtService, service.jwtService)
	}
}

func TestAuthService_DummyLogin(t *testing.T) {
	tests := []struct {
		name      string
		role      models.Role
		mockSetup func(*mockJWTService)
		wantToken string
		wantErr   string
	}{
		{
			name: "Valid employee role",
			role: models.RoleEmployee,
			mockSetup: func(m *mockJWTService) {
				m.On("GenerateToken", models.RoleEmployee).Return("employee-token", nil)
			},
			wantToken: "employee-token",
		},
		{
			name:      "Invalid role",
			role:      models.Role("invalid"),
			mockSetup: func(m *mockJWTService) {},
			wantErr:   "invalid role: invalid",
		},
		{
			name: "JWT error",
			role: models.RoleEmployee,
			mockSetup: func(m *mockJWTService) {
				m.On("GenerateToken", models.RoleEmployee).Return("", errors.New("jwt error"))
			},
			wantErr: "jwt error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockUserStorage{}
			jwtService := &mockJWTService{}
			service := NewAuthService(store, jwtService)

			tt.mockSetup(jwtService)

			token, err := service.DummyLogin(tt.role)

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("Expected error %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				if token != "" {
					t.Errorf("Expected empty token, got %q", token)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if token != tt.wantToken {
				t.Errorf("Expected token %q, got %q", tt.wantToken, token)
			}

			jwtService.AssertExpectations(t)
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		role      models.Role
		mockSetup func(*mockUserStorage, *mockJWTService)
		wantToken string
		wantErr   string
	}{
		{
			name:     "Successful registration",
			email:    "test@example.com",
			password: "password123",
			role:     models.RoleEmployee,
			mockSetup: func(store *mockUserStorage, jwt *mockJWTService) {
				store.On("GetUserByEmail", "test@example.com").Return((*models.User)(nil), nil)
				jwt.On("GenerateToken", models.RoleEmployee).Return("employee-token", nil)
				store.On("CreateUser", mock.AnythingOfType("*models.User"), "password123").Return("user-1", nil)
			},
			wantToken: "employee-token",
		},
		{
			name:      "Empty email",
			email:     "",
			password:  "password123",
			role:      models.RoleEmployee,
			mockSetup: func(store *mockUserStorage, jwt *mockJWTService) {},
			wantErr:   "email and password are required",
		},
		{
			name:      "Invalid role",
			email:     "test@example.com",
			password:  "password123",
			role:      models.Role("invalid"),
			mockSetup: func(store *mockUserStorage, jwt *mockJWTService) {},
			wantErr:   "invalid role: invalid",
		},
		{
			name:     "User already exists",
			email:    "test@example.com",
			password: "password123",
			role:     models.RoleEmployee,
			mockSetup: func(store *mockUserStorage, jwt *mockJWTService) {
				store.On("GetUserByEmail", "test@example.com").Return(&models.User{ID: "user-1", Email: "test@example.com", Role: models.RoleEmployee}, nil)
			},
			wantErr: "user with this email already exists",
		},
		{
			name:     "JWT error",
			email:    "test@example.com",
			password: "password123",
			role:     models.RoleEmployee,
			mockSetup: func(store *mockUserStorage, jwt *mockJWTService) {
				store.On("GetUserByEmail", "test@example.com").Return((*models.User)(nil), nil)
				jwt.On("GenerateToken", models.RoleEmployee).Return("", errors.New("jwt error"))
			},
			wantErr: "failed to generate token: jwt error",
		},
		{
			name:     "Storage error",
			email:    "test@example.com",
			password: "password123",
			role:     models.RoleEmployee,
			mockSetup: func(store *mockUserStorage, jwt *mockJWTService) {
				store.On("GetUserByEmail", "test@example.com").Return((*models.User)(nil), nil)
				jwt.On("GenerateToken", models.RoleEmployee).Return("employee-token", nil)
				store.On("CreateUser", mock.AnythingOfType("*models.User"), "password123").Return("", errors.New("storage error"))
			},
			wantErr: "failed to create user: storage error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockUserStorage{}
			jwtService := &mockJWTService{}
			service := NewAuthService(store, jwtService)

			tt.mockSetup(store, jwtService)

			token, err := service.Register(tt.email, tt.password, tt.role)

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("Expected error %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				if token != "" {
					t.Errorf("Expected empty token, got %q", token)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if token != tt.wantToken {
				t.Errorf("Expected token %q, got %q", tt.wantToken, token)
			}

			store.AssertExpectations(t)
			jwtService.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		mockSetup func(*mockUserStorage, *mockJWTService)
		wantToken string
		wantErr   string
	}{
		{
			name:     "Successful login",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func(store *mockUserStorage, jwt *mockJWTService) {
				store.On("GetUserByEmail", "test@example.com").Return(&models.User{ID: "user-1", Email: "test@example.com", Role: models.RoleEmployee}, nil)
				store.On("CheckPassword", "user-1", "password123").Return(true)
				jwt.On("GenerateToken", models.RoleEmployee).Return("employee-token", nil)
			},
			wantToken: "employee-token",
		},
		{
			name:      "Empty email",
			email:     "",
			password:  "password123",
			mockSetup: func(store *mockUserStorage, jwt *mockJWTService) {},
			wantErr:   "email and password are required",
		},
		{
			name:     "User not found",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func(store *mockUserStorage, jwt *mockJWTService) {
				store.On("GetUserByEmail", "test@example.com").Return((*models.User)(nil), nil)
			},
			wantErr: "user not found",
		},
		{
			name:     "Invalid password",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func(store *mockUserStorage, jwt *mockJWTService) {
				store.On("GetUserByEmail", "test@example.com").Return(&models.User{ID: "user-1", Email: "test@example.com", Role: models.RoleEmployee}, nil)
				store.On("CheckPassword", "user-1", "password123").Return(false)
			},
			wantErr: "invalid password",
		},
		{
			name:     "Storage error",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func(store *mockUserStorage, jwt *mockJWTService) {
				store.On("GetUserByEmail", "test@example.com").Return((*models.User)(nil), errors.New("storage error"))
			},
			wantErr: "failed to find user: storage error",
		},
		{
			name:     "JWT error",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func(store *mockUserStorage, jwt *mockJWTService) {
				store.On("GetUserByEmail", "test@example.com").Return(&models.User{ID: "user-1", Email: "test@example.com", Role: models.RoleEmployee}, nil)
				store.On("CheckPassword", "user-1", "password123").Return(true)
				jwt.On("GenerateToken", models.RoleEmployee).Return("", errors.New("jwt error"))
			},
			wantErr: "failed to generate token: jwt error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockUserStorage{}
			jwtService := &mockJWTService{}
			service := NewAuthService(store, jwtService)

			tt.mockSetup(store, jwtService)

			token, err := service.Login(tt.email, tt.password)

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("Expected error %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				if token != "" {
					t.Errorf("Expected empty token, got %q", token)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if token != tt.wantToken {
				t.Errorf("Expected token %q, got %q", tt.wantToken, token)
			}

			store.AssertExpectations(t)
			jwtService.AssertExpectations(t)
		})
	}
}
