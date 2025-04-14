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

// Мок для PVZServiceInterface
type mockPVZService struct {
	mock.Mock
}

func (m *mockPVZService) CreatePVZ(pvz *models.PVZ) (string, error) {
	args := m.Called(pvz)
	return args.String(0), args.Error(1)
}

func (m *mockPVZService) GetPVZsWithDetails(startDate, endDate string, page, pageSize int) ([]*models.PVZWithDetails, error) {
	args := m.Called(startDate, endDate, page, pageSize)
	return args.Get(0).([]*models.PVZWithDetails), args.Error(1)
}

// Тесты для PvzHandler (POST)
func TestPvzHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           interface{}
		body           interface{}
		mockSetup      func(*mockPVZService)
		expectedStatus int
		expectedBody   models.Error
		expectedPVZ    *models.PVZ
	}{
		{
			name: "Success",
			role: models.RoleModerator,
			body: map[string]string{"city": string(models.CitySaintPetersburg)},
			mockSetup: func(m *mockPVZService) {
				m.On("CreatePVZ", mock.AnythingOfType("*models.PVZ")).Return("123e4567-e89b-12d3-a456-426614174000", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedPVZ: &models.PVZ{
				ID:   "123e4567-e89b-12d3-a456-426614174000",
				City: models.CitySaintPetersburg,
			},
		},
		{
			name:           "Forbidden_NonModerator",
			role:           models.RoleEmployee,
			body:           map[string]string{"city": "Moscow"},
			mockSetup:      func(m *mockPVZService) {},
			expectedStatus: http.StatusForbidden,
			expectedBody:   models.Error{Message: "Only moderators can create PVZ"},
		},
		{
			name:           "Invalid_JSON",
			role:           models.RoleModerator,
			body:           "invalid json",
			mockSetup:      func(m *mockPVZService) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   models.Error{Message: "json: cannot unmarshal string into Go value of type struct { City models.City \"json:\\\"city\\\" binding:\\\"required\\\"\" }"},
		},
		{
			name:           "Missing_City",
			role:           models.RoleModerator,
			body:           map[string]string{},
			mockSetup:      func(m *mockPVZService) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   models.Error{Message: "Key: 'City' Error:Field validation for 'City' failed on the 'required' tag"},
		},
		{
			name: "Invalid_City",
			role: models.RoleModerator,
			body: map[string]string{"city": "InvalidCity"},
			mockSetup: func(m *mockPVZService) {
				m.On("CreatePVZ", mock.AnythingOfType("*models.PVZ")).Return("", errors.New("invalid city"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   models.Error{Message: "invalid city"},
		},
		{
			name: "Service_Error",
			role: models.RoleModerator,
			body: map[string]string{"city": "Moscow"},
			mockSetup: func(m *mockPVZService) {
				m.On("CreatePVZ", mock.AnythingOfType("*models.PVZ")).Return("", errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   models.Error{Message: "database error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка мока
			mockService := &mockPVZService{}
			tt.mockSetup(mockService)
			handler := NewPVZHandler(mockService, nil)

			// Подготовка запроса
			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Настройка контекста
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Set("role", tt.role)

			// Выполнение хендлера
			handler.PvzHandler(ctx)

			// Проверка
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedPVZ != nil {
				var response models.PVZ
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPVZ.ID, response.ID)
				assert.Equal(t, tt.expectedPVZ.City, response.City)
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

// Тесты для PvzGetHandler (GET)
func TestPvzGetHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           interface{}
		queryParams    map[string]string
		mockSetup      func(*mockPVZService)
		expectedStatus int
		expectedBody   string // JSON-строка для сравнения
	}{
		{
			name: "Success_Employee",
			role: models.RoleEmployee,
			queryParams: map[string]string{
				"startDate": "2023-01-01",
				"endDate":   "2023-12-31",
				"page":      "1",
				"limit":     "10",
			},
			mockSetup: func(m *mockPVZService) {
				m.On("GetPVZsWithDetails", "2023-01-01", "2023-12-31", 1, 10).Return([]*models.PVZWithDetails{
					{
						PVZ: &models.PVZ{
							ID:               "123e4567-e89b-12d3-a456-426614174000",
							RegistrationDate: "",
							City:             models.CityMoscow,
						},
						Receptions: []*models.ReceptionWithProducts{
							{
								Reception: &models.Reception{
									ID:       "456e7890-e89b-12d3-a456-426614174000",
									DateTime: "",
									PVZID:    "123e4567-e89b-12d3-a456-426614174000",
									Status:   models.ReceptionStatusClose,
								},
								Products: []*models.Product{
									{
										ID:          "789e1234-e89b-12d3-a456-426614174000",
										DateTime:    "",
										Type:        "электроника",
										ReceptionID: "",
									},
								},
							},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `[
                {
                    "pvz": {
                        "id": "123e4567-e89b-12d3-a456-426614174000",
                        "registrationDate": "",
                        "city": "Москва"
                    },
                    "receptions": [
                        {
                            "reception": {
                                "id": "456e7890-e89b-12d3-a456-426614174000",
                                "dateTime": "",
                                "pvzId": "123e4567-e89b-12d3-a456-426614174000",
                                "status": "close"
                            },
                            "products": [
                                {
                                    "id": "789e1234-e89b-12d3-a456-426614174000",
                                    "dateTime": "",
                                    "type": "электроника",
                                    "receptionId": ""
                                }
                            ]
                        }
                    ]
                }
            ]`,
		},
		{
			name: "Success_Moderator_Empty",
			role: models.RoleModerator,
			queryParams: map[string]string{
				"page":  "2",
				"limit": "20",
			},
			mockSetup: func(m *mockPVZService) {
				m.On("GetPVZsWithDetails", "", "", 2, 20).Return([]*models.PVZWithDetails{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[]`,
		},
		{
			name:           "Forbidden_Unauthorized",
			role:           "user",
			queryParams:    map[string]string{},
			mockSetup:      func(m *mockPVZService) {},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"message": "Only employees and moderators can access PVZ list"}`,
		},
		{
			name: "Invalid_Page",
			role: models.RoleEmployee,
			queryParams: map[string]string{
				"page": "-1",
			},
			mockSetup: func(m *mockPVZService) {
				m.On("GetPVZsWithDetails", "", "", 1, 10).Return([]*models.PVZWithDetails{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[]`,
		},
		{
			name: "Invalid_Limit",
			role: models.RoleEmployee,
			queryParams: map[string]string{
				"limit": "50",
			},
			mockSetup: func(m *mockPVZService) {
				m.On("GetPVZsWithDetails", "", "", 1, 10).Return([]*models.PVZWithDetails{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[]`,
		},
		{
			name: "Service_Error",
			role: models.RoleEmployee,
			queryParams: map[string]string{
				"startDate": "2023-01-01",
				"endDate":   "2023-12-31",
			},
			mockSetup: func(m *mockPVZService) {
				m.On("GetPVZsWithDetails", "2023-01-01", "2023-12-31", 1, 10).Return([]*models.PVZWithDetails(nil), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message": "database error"}`,
		},
		{
			name: "Nil_Details",
			role: models.RoleEmployee,
			queryParams: map[string]string{
				"page":  "1",
				"limit": "10",
			},
			mockSetup: func(m *mockPVZService) {
				m.On("GetPVZsWithDetails", "", "", 1, 10).Return([]*models.PVZWithDetails{
					nil,
					{
						PVZ:        nil,
						Receptions: []*models.ReceptionWithProducts{},
					},
					{
						PVZ: &models.PVZ{
							ID:               "123e4567-e89b-12d3-a456-426614174000",
							RegistrationDate: "",
							City:             models.CityMoscow,
						},
						Receptions: []*models.ReceptionWithProducts{
							nil,
							{
								Reception: nil,
								Products:  []*models.Product{},
							},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `[
                {
                    "pvz": {
                        "id": "123e4567-e89b-12d3-a456-426614174000",
                        "registrationDate": "",
                        "city": "Москва"
                    },
                    "receptions": []
                }
            ]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка мока
			mockService := &mockPVZService{}
			tt.mockSetup(mockService)
			handler := NewPVZHandler(mockService, nil)

			// Подготовка запроса
			req, _ := http.NewRequest(http.MethodGet, "/pvz", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
			w := httptest.NewRecorder()

			// Настройка контекста
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Set("role", tt.role)

			// Выполнение хендлера
			handler.PvzGetHandler(ctx)

			// Проверка
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Сравнение тела ответа как JSON
			var expectedJSON, actualJSON interface{}
			if tt.expectedBody != "" {
				err := json.Unmarshal([]byte(tt.expectedBody), &expectedJSON)
				assert.NoError(t, err, "Failed to unmarshal expected body")
				err = json.Unmarshal(w.Body.Bytes(), &actualJSON)
				assert.NoError(t, err, "Failed to unmarshal actual body")
				assert.Equal(t, expectedJSON, actualJSON, "Response body mismatch")
			} else {
				assert.Empty(t, w.Body.String(), "Expected empty body")
			}

			// Проверка ожиданий мока
			mockService.AssertExpectations(t)
		})
	}
}
