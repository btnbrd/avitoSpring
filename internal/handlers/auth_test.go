package handlers

import (
	"avitoSpring/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) DummyLogin(role models.Role) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func (m *mockAuthService) Register(email, password string, role models.Role) (string, error) {
	args := m.Called(email, password, role)
	return args.String(0), args.Error(1)
}

func (m *mockAuthService) Login(email, password string) (string, error) {
	args := m.Called(email, password)
	return args.String(0), args.Error(1)
}

func TestNewAuthHandler(t *testing.T) {
	authService := &mockAuthService{}
	handler := NewAuthHandler(authService)

	if handler == nil {
		t.Fatal("Expected non-nil AuthHandler")
	}
	if handler.authService != authService {
		t.Errorf("Expected authService %v, got %v", authService, handler.authService)
	}
}

func TestAuthHandler_RegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       interface{}
		mockSetup  func(*mockAuthService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "Successful registration",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
				"role":     "employee",
			},
			mockSetup: func(m *mockAuthService) {
				m.On("Register", "test@example.com", "password123", models.Role("employee")).Return("employee-token", nil)
			},
			wantStatus: http.StatusCreated,
			wantBody:   `{"token":"employee-token"}`,
		},
		{
			name: "Missing email",
			body: map[string]string{
				"password": "password123",
				"role":     "employee",
			},
			mockSetup:  func(m *mockAuthService) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"Key: 'Email' Error:Field validation for 'Email' failed on the 'required' tag"}`,
		},
		{
			name: "Invalid email",
			body: map[string]string{
				"email":    "invalid-email",
				"password": "password123",
				"role":     "employee",
			},
			mockSetup:  func(m *mockAuthService) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"Key: 'Email' Error:Field validation for 'Email' failed on the 'email' tag"}`,
		},
		{
			name: "AuthService error",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
				"role":     "employee",
			},
			mockSetup: func(m *mockAuthService) {
				m.On("Register", "test@example.com", "password123", models.Role("employee")).Return("", errors.New("user already exists"))
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"user already exists"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := &mockAuthService{}
			handler := NewAuthHandler(authService)

			tt.mockSetup(authService)

			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.RegisterHandler(c)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if strings.TrimSpace(w.Body.String()) != tt.wantBody {
				t.Errorf("Expected body %q, got %q", tt.wantBody, w.Body.String())
			}

			authService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_LoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       interface{}
		mockSetup  func(*mockAuthService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "Successful login",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockSetup: func(m *mockAuthService) {
				m.On("Login", "test@example.com", "password123").Return("employee-token", nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"token":"employee-token"}`,
		},
		{
			name: "Missing password",
			body: map[string]string{
				"email": "test@example.com",
			},
			mockSetup:  func(m *mockAuthService) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"Key: 'Password' Error:Field validation for 'Password' failed on the 'required' tag"}`,
		},
		{
			name: "AuthService error",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockSetup: func(m *mockAuthService) {
				m.On("Login", "test@example.com", "password123").Return("", errors.New("invalid password"))
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"invalid password"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := &mockAuthService{}
			handler := NewAuthHandler(authService)

			tt.mockSetup(authService)

			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.LoginHandler(c)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if strings.TrimSpace(w.Body.String()) != tt.wantBody {
				t.Errorf("Expected body %q, got %q", tt.wantBody, w.Body.String())
			}

			authService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_DummyLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       interface{}
		mockSetup  func(*mockAuthService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "Successful dummy login",
			body: map[string]string{
				"role": "employee",
			},
			mockSetup: func(m *mockAuthService) {
				m.On("DummyLogin", models.Role("employee")).Return("employee-token", nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"token":"employee-token"}`,
		},
		{
			name:       "Missing role",
			body:       map[string]string{},
			mockSetup:  func(m *mockAuthService) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"role is required"}`,
		},
		{
			name: "AuthService error",
			body: map[string]string{
				"role": "employee",
			},
			mockSetup: func(m *mockAuthService) {
				m.On("DummyLogin", models.Role("employee")).Return("", errors.New("invalid role"))
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"invalid role"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := &mockAuthService{}
			handler := NewAuthHandler(authService)

			tt.mockSetup(authService)

			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPost, "/dummy-login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.DummyLoginHandler(c)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if strings.TrimSpace(w.Body.String()) != tt.wantBody {
				t.Errorf("Expected body %q, got %q", tt.wantBody, w.Body.String())
			}

			authService.AssertExpectations(t)
		})
	}
}
