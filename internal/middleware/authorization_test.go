package middleware

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/services"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubJWTService struct {
	validateErr bool
	role        models.Role
}

func (s *stubJWTService) GenerateToken(role models.Role) (string, error) {
	return string(role) + "-token", nil
}

func (s *stubJWTService) ValidateToken(tokenString string) (models.Role, error) {
	if s.validateErr {
		return "", errors.New("invalid token")
	}
	if tokenString == "" {
		return "", errors.New("token is empty")
	}
	return s.role, nil
}

var _ services.JWTServiceInterface = (*stubJWTService)(nil)

func TestNewAuthenticator(t *testing.T) {
	jwtService := &stubJWTService{}
	authenticator := NewAuthorizer(jwtService)

	if authenticator == nil {
		t.Fatal("Expected non-nil Authenticator")
	}
	if authenticator.jwtService != jwtService {
		t.Errorf("Expected jwtService %v, got %v", jwtService, authenticator.jwtService)
	}
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		header      map[string]string
		validateErr bool
		role        models.Role
		wantStatus  int
		wantBody    string
		wantRole    models.Role
	}{
		{
			name: "Successful authentication",
			header: map[string]string{
				"Authorization": "Bearer valid-token",
			},
			role:       models.RoleEmployee,
			wantStatus: http.StatusOK,
			wantBody:   "",
			wantRole:   models.RoleEmployee,
		},
		{
			name:       "Missing Authorization header",
			header:     map[string]string{},
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"message":"Authorization header is required"}`,
			wantRole:   "",
		},
		{
			name: "Invalid Authorization format",
			header: map[string]string{
				"Authorization": "Invalid-token",
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"message":"Authorization header must start with 'Bearer '"}`,
			wantRole:   "",
		},
		{
			name: "Invalid token",
			header: map[string]string{
				"Authorization": "Bearer invalid-token",
			},
			validateErr: true,
			wantStatus:  http.StatusUnauthorized,
			wantBody:    `{"message":"invalid token"}`,
			wantRole:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtService := &stubJWTService{
				validateErr: tt.validateErr,
				role:        tt.role,
			}
			authenticator := NewAuthorizer(jwtService)

			var capturedRole models.Role
			handler := func(c *gin.Context) {
				if role, exists := c.Get("role"); exists {
					var ok bool
					if capturedRole, ok = role.(models.Role); !ok {
						t.Errorf("Expected role to be models.Role, got %T", role)
					}
				}
				c.String(http.StatusOK, "")
			}

			r := gin.New()
			r.Use(authenticator.AuthMiddleware())
			r.GET("/test", handler)

			// Создаем запрос
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			for k, v := range tt.header {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if strings.TrimSpace(w.Body.String()) != tt.wantBody {
				t.Errorf("Expected body %q, got %q", tt.wantBody, w.Body.String())
			}

			if capturedRole != tt.wantRole {
				t.Errorf("Expected role %q in context, got %q", tt.wantRole, capturedRole)
			}
		})
	}
}
