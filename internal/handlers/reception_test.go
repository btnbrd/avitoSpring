package handlers

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/services"
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

// Мок для ReceptionServiceInterface
type mockReceptionService struct {
	mock.Mock
}

var _ services.ReceptionServiceInterface = (*mockReceptionService)(nil)

func (m *mockReceptionService) CreateReception(reception *models.Reception) (string, error) {
	args := m.Called(reception)
	return args.String(0), args.Error(1)
}

func (m *mockReceptionService) CloseLastReception(pvzID string) error {
	args := m.Called(pvzID)
	return args.Error(0)
}

func (m *mockReceptionService) GetLastReceptionByPVZID(pvzID string) (*models.Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(*models.Reception), args.Error(1)
}

func TestReceptionHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		role              interface{}
		body              interface{}
		mockSetup         func(*mockReceptionService)
		expectedStatus    int
		expectedBody      models.Error
		expectedReception *models.Reception
	}{
		{
			name: "Success",
			role: models.RoleEmployee,
			body: map[string]string{"pvzId": "d5848505-3d33-47bc-8e65-48837815f623"},
			mockSetup: func(m *mockReceptionService) {
				m.On("CreateReception", mock.AnythingOfType("*models.Reception")).Return("123e4567-e89b-12d3-a456-426614174000", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedReception: &models.Reception{
				ID:    "123e4567-e89b-12d3-a456-426614174000",
				PVZID: "d5848505-3d33-47bc-8e65-48837815f623",
			},
		},
		{
			name:           "Forbidden_NonEmployee",
			role:           models.RoleModerator,
			body:           map[string]string{"pvzId": "d5848505-3d33-47bc-8e65-48837815f623"},
			mockSetup:      func(m *mockReceptionService) {},
			expectedStatus: http.StatusForbidden,
			expectedBody:   models.Error{Message: "Only employees can create receptions"},
		},
		{
			name:           "Invalid_JSON",
			role:           models.RoleEmployee,
			body:           "invalid json",
			mockSetup:      func(m *mockReceptionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.Error{Message: "json: cannot unmarshal string into Go value of type struct { PVZID string \"json:\\\"pvzId\\\" binding:\\\"required,uuid\\\"\" }"},
		},
		{
			name:           "Missing_PVZID",
			role:           models.RoleEmployee,
			body:           map[string]string{},
			mockSetup:      func(m *mockReceptionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.Error{Message: "Key: 'PVZID' Error:Field validation for 'PVZID' failed on the 'required' tag"},
		},
		{
			name:           "Invalid_PVZID",
			role:           models.RoleEmployee,
			body:           map[string]string{"pvzId": "invalid-uuid"},
			mockSetup:      func(m *mockReceptionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.Error{Message: "Key: 'PVZID' Error:Field validation for 'PVZID' failed on the 'uuid' tag"},
		},
		{
			name: "Service_Error",
			role: models.RoleEmployee,
			body: map[string]string{"pvzId": "d5848505-3d33-47bc-8e65-48837815f623"},
			mockSetup: func(m *mockReceptionService) {
				m.On("CreateReception", mock.AnythingOfType("*models.Reception")).Return("", errors.New("database error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.Error{Message: "database error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка мока
			mockService := &mockReceptionService{}
			tt.mockSetup(mockService)
			handler := NewReceptionHandler(mockService, nil) // AuthHandler не нужен

			// Подготовка запроса
			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPost, "/receptions", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Настройка контекста
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Set("role", tt.role)

			// Выполнение хендлера
			handler.ReceptionHandler(ctx)

			// Проверка
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedReception != nil {
				var response models.Reception
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedReception.ID, response.ID)
				assert.Equal(t, tt.expectedReception.PVZID, response.PVZID)
			} else {
				var response models.Error
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, tt.expectedBody.Message)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// Тесты для CloseLastReceptionHandler
func TestCloseLastReceptionHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		role              interface{}
		pvzID             string
		mockSetup         func(*mockReceptionService)
		expectedStatus    int
		expectedBody      models.Error
		expectedReception *models.Reception
	}{
		{
			name:  "Success",
			role:  models.RoleEmployee,
			pvzID: "d5848505-3d33-47bc-8e65-48837815f623",
			mockSetup: func(m *mockReceptionService) {
				m.On("CloseLastReception", "d5848505-3d33-47bc-8e65-48837815f623").Return(nil)
				m.On("GetLastReceptionByPVZID", "d5848505-3d33-47bc-8e65-48837815f623").Return(&models.Reception{
					ID:     "123e4567-e89b-12d3-a456-426614174000",
					PVZID:  "d5848505-3d33-47bc-8e65-48837815f623",
					Status: models.ReceptionStatusClose,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedReception: &models.Reception{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				PVZID:  "d5848505-3d33-47bc-8e65-48837815f623",
				Status: models.ReceptionStatusClose,
			},
		},
		{
			name:           "Forbidden_NonEmployee",
			role:           models.RoleModerator,
			pvzID:          "d5848505-3d33-47bc-8e65-48837815f623",
			mockSetup:      func(m *mockReceptionService) {},
			expectedStatus: http.StatusForbidden,
			expectedBody:   models.Error{Message: "Only employees can close receptions"},
		},
		{
			name:           "Missing_PVZID",
			role:           models.RoleEmployee,
			pvzID:          "",
			mockSetup:      func(m *mockReceptionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.Error{Message: "pvzId is required"},
		},
		{
			name:  "CloseReception_Error",
			role:  models.RoleEmployee,
			pvzID: "d5848505-3d33-47bc-8e65-48837815f623",
			mockSetup: func(m *mockReceptionService) {
				m.On("CloseLastReception", "d5848505-3d33-47bc-8e65-48837815f623").Return(errors.New("database error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.Error{Message: "database error"},
		},
		{
			name:  "No_Reception_Found",
			role:  models.RoleEmployee,
			pvzID: "d5848505-3d33-47bc-8e65-48837815f623",
			mockSetup: func(m *mockReceptionService) {
				m.On("CloseLastReception", "d5848505-3d33-47bc-8e65-48837815f623").Return(nil)
				m.On("GetLastReceptionByPVZID", "d5848505-3d33-47bc-8e65-48837815f623").Return((*models.Reception)(nil), nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.Error{Message: "no reception found for PVZ"},
		},
		{
			name:  "GetReception_Error",
			role:  models.RoleEmployee,
			pvzID: "d5848505-3d33-47bc-8e65-48837815f623",
			mockSetup: func(m *mockReceptionService) {
				m.On("CloseLastReception", "d5848505-3d33-47bc-8e65-48837815f623").Return(nil)
				m.On("GetLastReceptionByPVZID", "d5848505-3d33-47bc-8e65-48837815f623").Return((*models.Reception)(nil), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   models.Error{Message: "database error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка мока
			mockService := &mockReceptionService{}
			tt.mockSetup(mockService)
			handler := NewReceptionHandler(mockService, nil) // AuthHandler не нужен

			// Подготовка запроса
			req, _ := http.NewRequest(http.MethodPost, "/pvz/"+tt.pvzID+"/close_last_reception", nil)
			w := httptest.NewRecorder()

			// Настройка контекста
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Set("role", tt.role)
			ctx.Params = gin.Params{{Key: "pvzId", Value: tt.pvzID}}

			// Выполнение хендлера
			handler.CloseLastReceptionHandler(ctx)

			// Проверка
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedReception != nil {
				var response models.Reception
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedReception.ID, response.ID)
				assert.Equal(t, tt.expectedReception.PVZID, response.PVZID)
				assert.Equal(t, tt.expectedReception.Status, response.Status)
			} else {
				var response models.Error
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.Message, response.Message)
			}
			mockService.AssertExpectations(t)
		})
	}
}
