package services

import (
	"avitoSpring/internal/models"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мок для ReceptionStorage
type mockReceptionStorage struct {
	mock.Mock
}

func (m *mockReceptionStorage) CreateReception(reception *models.Reception) (string, error) {
	args := m.Called(reception)
	return args.String(0), args.Error(1)
}

func (m *mockReceptionStorage) HasOpenReception(pvzID string) (bool, error) {
	args := m.Called(pvzID)
	return args.Bool(0), args.Error(1)
}

func (m *mockReceptionStorage) CloseLastReception(pvzID string) error {
	args := m.Called(pvzID)
	return args.Error(0)
}

func (m *mockReceptionStorage) GetLastReceptionByPVZID(pvzID string) (*models.Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(*models.Reception), args.Error(1)
}

func TestReceptionService_CreateReception(t *testing.T) {
	tests := []struct {
		name          string
		reception     *models.Reception
		mockSetup     func(*mockReceptionStorage)
		expectedID    string
		expectedError string
	}{
		{
			name: "Success",
			reception: &models.Reception{
				PVZID:    "pvz1",
				Status:   models.ReceptionStatusInProgress,
				DateTime: "2023-01-01T12:00:00Z",
			},
			mockSetup: func(rs *mockReceptionStorage) {
				rs.On("HasOpenReception", "pvz1").Return(false, nil)
				rs.On("CreateReception", mock.AnythingOfType("*models.Reception")).Return("rec1", nil)
			},
			expectedID:    "rec1",
			expectedError: "",
		},
		{
			name: "Open_Reception_Exists",
			reception: &models.Reception{
				PVZID:    "pvz1",
				Status:   models.ReceptionStatusInProgress,
				DateTime: "2023-01-01T12:00:00Z",
			},
			mockSetup: func(rs *mockReceptionStorage) {
				rs.On("HasOpenReception", "pvz1").Return(true, nil)
			},
			expectedID:    "",
			expectedError: "an open reception already exists for PVZ pvz1",
		},
		{
			name: "Empty_DateTime",
			reception: &models.Reception{
				PVZID:  "pvz1",
				Status: models.ReceptionStatusInProgress,
			},
			mockSetup: func(rs *mockReceptionStorage) {
				rs.On("HasOpenReception", "pvz1").Return(false, nil)
				rs.On("CreateReception", mock.AnythingOfType("*models.Reception")).Return("rec2", nil)
			},
			expectedID:    "rec2",
			expectedError: "",
		},
		{
			name: "Empty_Status",
			reception: &models.Reception{
				PVZID:    "pvz1",
				DateTime: "2023-01-01T12:00:00Z",
			},
			mockSetup: func(rs *mockReceptionStorage) {
				rs.On("HasOpenReception", "pvz1").Return(false, nil)
				rs.On("CreateReception", mock.AnythingOfType("*models.Reception")).Return("rec3", nil)
			},
			expectedID:    "rec3",
			expectedError: "",
		},
		{
			name: "Invalid_Status",
			reception: &models.Reception{
				PVZID:    "pvz1",
				Status:   "invalid",
				DateTime: "2023-01-01T12:00:00Z",
			},
			mockSetup: func(rs *mockReceptionStorage) {
				rs.On("HasOpenReception", "pvz1").Return(false, nil)
			},
			expectedID:    "",
			expectedError: "invalid status: invalid",
		},
		{
			name: "Empty_PVZID",
			reception: &models.Reception{
				Status:   models.ReceptionStatusInProgress,
				DateTime: "2023-01-01T12:00:00Z",
			},
			mockSetup: func(rs *mockReceptionStorage) {
				// Нет вызовов HasOpenReception, так как проверка PVZID происходит раньше
			},
			expectedID:    "",
			expectedError: "pvzId is required",
		},
		{
			name: "HasOpenReception_Error",
			reception: &models.Reception{
				PVZID:    "pvz1",
				Status:   models.ReceptionStatusInProgress,
				DateTime: "2023-01-01T12:00:00Z",
			},
			mockSetup: func(rs *mockReceptionStorage) {
				rs.On("HasOpenReception", "pvz1").Return(false, errors.New("storage error"))
			},
			expectedID:    "",
			expectedError: "failed to check open receptions: storage error",
		},
		{
			name: "CreateReception_Error",
			reception: &models.Reception{
				PVZID:    "pvz1",
				Status:   models.ReceptionStatusInProgress,
				DateTime: "2023-01-01T12:00:00Z",
			},
			mockSetup: func(rs *mockReceptionStorage) {
				rs.On("HasOpenReception", "pvz1").Return(false, nil)
				rs.On("CreateReception", mock.AnythingOfType("*models.Reception")).Return("", errors.New("storage error"))
			},
			expectedID:    "",
			expectedError: "storage error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка мока
			receptionStorage := &mockReceptionStorage{}
			tt.mockSetup(receptionStorage)

			// Создание сервиса
			service := NewReceptionService(receptionStorage)

			// Выполнение метода
			id, err := service.CreateReception(tt.reception)

			// Проверка результатов
			assert.Equal(t, tt.expectedID, id)
			if tt.expectedError == "" {
				assert.NoError(t, err)
				if tt.name == "Empty_DateTime" {
					// Проверяем, что DateTime установлена
					assert.NotEmpty(t, tt.reception.DateTime)
					_, parseErr := time.Parse(time.RFC3339, tt.reception.DateTime)
					assert.NoError(t, parseErr)
				}
				if tt.name == "Empty_Status" {
					// Проверяем, что Status установлен в in_progress
					assert.Equal(t, models.ReceptionStatusInProgress, tt.reception.Status)
				}
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}

			// Проверка вызовов мока
			receptionStorage.AssertExpectations(t)
		})
	}
}

func TestReceptionService_GetLastReceptionByPVZID(t *testing.T) {
	tests := []struct {
		name           string
		pvzID          string
		mockSetup      func(*mockReceptionStorage)
		expectedResult *models.Reception
		expectedError  string
	}{
		{
			name:  "Success",
			pvzID: "pvz1",
			mockSetup: func(rs *mockReceptionStorage) {
				rs.On("GetLastReceptionByPVZID", "pvz1").Return(&models.Reception{
					ID:       "rec1",
					PVZID:    "pvz1",
					Status:   models.ReceptionStatusInProgress,
					DateTime: "2023-01-01T12:00:00Z",
				}, nil)
			},
			expectedResult: &models.Reception{
				ID:       "rec1",
				PVZID:    "pvz1",
				Status:   models.ReceptionStatusInProgress,
				DateTime: "2023-01-01T12:00:00Z",
			},
			expectedError: "",
		},
		{
			name:           "Empty_PVZID",
			pvzID:          "",
			mockSetup:      func(rs *mockReceptionStorage) {},
			expectedResult: nil,
			expectedError:  "pvzID is required",
		},
		{
			name:  "Storage_Error",
			pvzID: "pvz1",
			mockSetup: func(rs *mockReceptionStorage) {
				rs.On("GetLastReceptionByPVZID", "pvz1").Return((*models.Reception)(nil), errors.New("storage error"))
			},
			expectedResult: nil,
			expectedError:  "storage error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка мока
			receptionStorage := &mockReceptionStorage{}
			tt.mockSetup(receptionStorage)

			// Создание сервиса
			service := NewReceptionService(receptionStorage)

			// Выполнение метода
			result, err := service.GetLastReceptionByPVZID(tt.pvzID)

			// Проверка результатов
			assert.Equal(t, tt.expectedResult, result)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}

			// Проверка вызовов мока
			receptionStorage.AssertExpectations(t)
		})
	}
}

func TestReceptionService_CloseLastReception(t *testing.T) {
	tests := []struct {
		name          string
		pvzID         string
		mockSetup     func(*mockReceptionStorage)
		expectedError string
	}{
		{
			name:  "Success",
			pvzID: "pvz1",
			mockSetup: func(rs *mockReceptionStorage) {
				rs.On("CloseLastReception", "pvz1").Return(nil)
			},
			expectedError: "",
		},
		{
			name:          "Empty_PVZID",
			pvzID:         "",
			mockSetup:     func(rs *mockReceptionStorage) {},
			expectedError: "pvzID is required",
		},
		{
			name:  "Storage_Error",
			pvzID: "pvz1",
			mockSetup: func(rs *mockReceptionStorage) {
				rs.On("CloseLastReception", "pvz1").Return(errors.New("storage error"))
			},
			expectedError: "storage error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка мока
			receptionStorage := &mockReceptionStorage{}
			tt.mockSetup(receptionStorage)

			// Создание сервиса
			service := NewReceptionService(receptionStorage)

			// Выполнение метода
			err := service.CloseLastReception(tt.pvzID)

			// Проверка результатов
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}

			// Проверка вызовов мока
			receptionStorage.AssertExpectations(t)
		})
	}
}
