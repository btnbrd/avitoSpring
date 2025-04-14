package handlers

import (
	"avitoSpring/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мок для ProductServiceInterface
type mockProductService struct {
	mock.Mock
}

func (m *mockProductService) CreateProduct(product *models.Product, pvzID string) (string, error) {
	args := m.Called(product, pvzID)
	return args.String(0), args.Error(1)
}

func (m *mockProductService) DeleteLastProduct(pvzID string) error {
	args := m.Called(pvzID)
	return args.Error(0)
}

func TestProductHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           interface{}
		body           interface{}
		mockSetup      func(*mockProductService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			role: models.RoleEmployee,
			body: map[string]string{
				"type":  "electronics",
				"pvzId": "123e4567-e89b-12d3-a456-426614174000",
			},
			mockSetup: func(m *mockProductService) {
				m.On("CreateProduct", mock.AnythingOfType("*models.Product"), "123e4567-e89b-12d3-a456-426614174000").
					Return("prod1", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id": "prod1", "type": "electronics", "dateTime": "", "receptionId": ""}`,
		},
		{
			name:           "Forbidden_NonEmployee",
			role:           "moderator",
			body:           map[string]string{"type": "electronics", "pvzId": "123e4567-e89b-12d3-a456-426614174000"},
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"message": "Only employees can create products"}`,
		},
		{
			name:           "Invalid_JSON",
			role:           models.RoleEmployee,
			body:           "invalid json",
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,

			expectedBody: `{"message": "json: cannot unmarshal string into Go value of type struct { Type models.ProductType \"json:\\\"type\\\" binding:\\\"required\\\"\"; PVZID string \"json:\\\"pvzId\\\" binding:\\\"required,uuid\\\"\" }"}`,
		},
		{
			name:           "Missing_Type",
			role:           models.RoleEmployee,
			body:           map[string]string{"pvzId": "123e4567-e89b-12d3-a456-426614174000"},
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message": "Key: 'Type' Error:Field validation for 'Type' failed on the 'required' tag"}`,
		},
		{
			name:           "Missing_PVZID",
			role:           models.RoleEmployee,
			body:           map[string]string{"type": "electronics"},
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message": "Key: 'PVZID' Error:Field validation for 'PVZID' failed on the 'required' tag"}`,
		},
		{
			name:           "Invalid_PVZID",
			role:           models.RoleEmployee,
			body:           map[string]string{"type": "electronics", "pvzId": "invalid-uuid"},
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message": "Key: 'PVZID' Error:Field validation for 'PVZID' failed on the 'uuid' tag"}`,
		},
		{
			name: "Service_Error",
			role: models.RoleEmployee,
			body: map[string]string{
				"type":  "electronics",
				"pvzId": "123e4567-e89b-12d3-a456-426614174000",
			},
			mockSetup: func(m *mockProductService) {
				m.On("CreateProduct", mock.AnythingOfType("*models.Product"), "123e4567-e89b-12d3-a456-426614174000").
					Return("", errors.New("service error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message": "service error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка мока
			mockService := &mockProductService{}
			tt.mockSetup(mockService)
			handler := NewProductHandler(mockService, nil)

			// Подготовка запроса
			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Настройка контекста
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Set("role", tt.role)

			// Выполнение хендлера
			handler.ProductHandler(ctx)

			// Проверка
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

func TestDeleteLastProductHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           interface{}
		pvzID          string
		mockSetup      func(*mockProductService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "Success",
			role:  models.RoleEmployee,
			pvzID: "123e4567-e89b-12d3-a456-426614174000",
			mockSetup: func(m *mockProductService) {
				m.On("DeleteLastProduct", "123e4567-e89b-12d3-a456-426614174000").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "Missing_PVZID",
			role:           models.RoleEmployee,
			pvzID:          "",
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message": "pvzId is required"}`,
		},
		{
			name:           "Forbidden_NonEmployee",
			role:           "moderator",
			pvzID:          "123e4567-e89b-12d3-a456-426614174000",
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"message": "Only employees can delete products"}`,
		},
		{
			name:  "Service_Error",
			role:  models.RoleEmployee,
			pvzID: "123e4567-e89b-12d3-a456-426614174000",
			mockSetup: func(m *mockProductService) {
				m.On("DeleteLastProduct", "123e4567-e89b-12d3-a456-426614174000").Return(errors.New("service error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message": "service error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка мока
			mockService := &mockProductService{}
			tt.mockSetup(mockService)
			handler := NewProductHandler(mockService, nil)

			// Подготовка запроса
			req, _ := http.NewRequest(http.MethodDelete, "/products/"+tt.pvzID, nil)
			w := httptest.NewRecorder()

			// Настройка контекста
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Set("role", tt.role)
			if tt.pvzID != "" {
				ctx.Params = []gin.Param{{Key: "pvzId", Value: tt.pvzID}}
			}

			// Выполнение хендлера
			handler.DeleteLastProductHandler(ctx)

			// Проверка
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody == "" {
				assert.Empty(t, w.Body.String())
			} else {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}
