package services

import (
	"avitoSpring/internal/models"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мок для ProductStorage
type mockProductStorage struct {
	mock.Mock
}

func (m *mockProductStorage) CreateProduct(product *models.Product) (string, error) {
	args := m.Called(product)
	return args.String(0), args.Error(1)
}

func (m *mockProductStorage) DeleteLastProduct(pvzID string) error {
	args := m.Called(pvzID)
	return args.Error(0)
}

//// Мок для ReceptionStorage
//type mockReceptionStorage struct {
//	mock.Mock
//}
//
//func (m *mockReceptionStorage) CreateReception(reception *models.Reception) (string, error) {
//	args := m.Called(reception)
//	return args.String(0), args.Error(1)
//}
//
//func (m *mockReceptionStorage) HasOpenReception(pvzID string) (bool, error) {
//	args := m.Called(pvzID)
//	return args.Bool(0), args.Error(1)
//}
//
//func (m *mockReceptionStorage) CloseLastReception(pvzID string) error {
//	args := m.Called(pvzID)
//	return args.Error(0)
//}
//
//func (m *mockReceptionStorage) GetLastReceptionByPVZID(pvzID string) (*models.Reception, error) {
//	args := m.Called(pvzID)
//	return args.Get(0).(*models.Reception), args.Error(1)
//}

func TestProductService_CreateProduct(t *testing.T) {
	tests := []struct {
		name          string
		product       *models.Product
		pvzID         string
		mockSetup     func(*mockProductStorage, *mockReceptionStorage)
		expectedID    string
		expectedError string
	}{
		{
			name: "Success",
			product: &models.Product{
				Type:     models.ProductTypeElectronics,
				DateTime: "2023-01-01T12:00:00Z",
			},
			pvzID: "pvz1",
			mockSetup: func(ps *mockProductStorage, rs *mockReceptionStorage) {
				rs.On("GetLastReceptionByPVZID", "pvz1").Return(&models.Reception{
					ID:     "rec1",
					PVZID:  "pvz1",
					Status: models.ReceptionStatusInProgress,
				}, nil)
				ps.On("CreateProduct", mock.AnythingOfType("*models.Product")).Return("prod1", nil)
			},
			expectedID:    "prod1",
			expectedError: "",
		},
		{
			name: "Invalid_Product_Type",
			product: &models.Product{
				Type:     "invalid",
				DateTime: "2023-01-01T12:00:00Z",
			},
			pvzID: "pvz1",
			mockSetup: func(ps *mockProductStorage, rs *mockReceptionStorage) {
				// Нет вызовов мока, так как валидация типа происходит до обращения к хранилищу
			},
			expectedID:    "",
			expectedError: "invalid product type: invalid",
		},
		{
			name: "Empty_DateTime",
			product: &models.Product{
				Type: models.ProductTypeClothing,
			},
			pvzID: "pvz1",
			mockSetup: func(ps *mockProductStorage, rs *mockReceptionStorage) {
				rs.On("GetLastReceptionByPVZID", "pvz1").Return(&models.Reception{
					ID:     "rec1",
					PVZID:  "pvz1",
					Status: models.ReceptionStatusInProgress,
				}, nil)
				ps.On("CreateProduct", mock.AnythingOfType("*models.Product")).Return("prod2", nil)
			},
			expectedID:    "prod2",
			expectedError: "",
		},
		{
			name: "Reception_Storage_Error",
			product: &models.Product{
				Type:     models.ProductTypeFootwear,
				DateTime: "2023-01-01T12:00:00Z",
			},
			pvzID: "pvz1",
			mockSetup: func(ps *mockProductStorage, rs *mockReceptionStorage) {
				rs.On("GetLastReceptionByPVZID", "pvz1").Return((*models.Reception)(nil), errors.New("storage error"))
			},
			expectedID:    "",
			expectedError: "failed to get last reception: storage error",
		},
		{
			name: "No_Open_Reception",
			product: &models.Product{
				Type:     models.ProductTypeElectronics,
				DateTime: "2023-01-01T12:00:00Z",
			},
			pvzID: "pvz1",
			mockSetup: func(ps *mockProductStorage, rs *mockReceptionStorage) {
				rs.On("GetLastReceptionByPVZID", "pvz1").Return((*models.Reception)(nil), nil)
			},
			expectedID:    "",
			expectedError: "no open reception found for PVZ pvz1",
		},
		{
			name: "Closed_Reception",
			product: &models.Product{
				Type:     models.ProductTypeClothing,
				DateTime: "2023-01-01T12:00:00Z",
			},
			pvzID: "pvz1",
			mockSetup: func(ps *mockProductStorage, rs *mockReceptionStorage) {
				rs.On("GetLastReceptionByPVZID", "pvz1").Return(&models.Reception{
					ID:     "rec1",
					PVZID:  "pvz1",
					Status: models.ReceptionStatusClose,
				}, nil)
			},
			expectedID:    "",
			expectedError: "no open reception found for PVZ pvz1",
		},
		{
			name: "Product_Storage_Error",
			product: &models.Product{
				Type:     models.ProductTypeFootwear,
				DateTime: "2023-01-01T12:00:00Z",
			},
			pvzID: "pvz1",
			mockSetup: func(ps *mockProductStorage, rs *mockReceptionStorage) {
				rs.On("GetLastReceptionByPVZID", "pvz1").Return(&models.Reception{
					ID:     "rec1",
					PVZID:  "pvz1",
					Status: models.ReceptionStatusInProgress,
				}, nil)
				ps.On("CreateProduct", mock.AnythingOfType("*models.Product")).Return("", errors.New("storage error"))
			},
			expectedID:    "",
			expectedError: "storage error",
		},
		{
			name: "Empty_PVZID",
			product: &models.Product{
				Type:     models.ProductTypeElectronics,
				DateTime: "2023-01-01T12:00:00Z",
			},
			pvzID: "",
			mockSetup: func(ps *mockProductStorage, rs *mockReceptionStorage) {
				rs.On("GetLastReceptionByPVZID", "").Return((*models.Reception)(nil), errors.New("invalid pvzID"))
			},
			expectedID:    "",
			expectedError: "failed to get last reception: invalid pvzID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка моков
			productStorage := &mockProductStorage{}
			receptionStorage := &mockReceptionStorage{}
			tt.mockSetup(productStorage, receptionStorage)

			// Создание сервиса
			service := NewProductService(productStorage, receptionStorage)

			// Выполнение метода
			id, err := service.CreateProduct(tt.product, tt.pvzID)

			// Проверка результатов
			assert.Equal(t, tt.expectedID, id)
			if tt.expectedError == "" {
				assert.NoError(t, err)
				if tt.name == "Empty_DateTime" {
					// Проверяем, что DateTime установлена
					assert.NotEmpty(t, tt.product.DateTime)
					_, parseErr := time.Parse(time.RFC3339, tt.product.DateTime)
					assert.NoError(t, parseErr)
				}
				if tt.expectedID != "" {
					// Проверяем, что ReceptionID установлен
					assert.Equal(t, "rec1", tt.product.ReceptionID)
				}
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}

			// Проверка вызовов моков
			productStorage.AssertExpectations(t)
			receptionStorage.AssertExpectations(t)
		})
	}
}

func TestProductService_DeleteLastProduct(t *testing.T) {
	tests := []struct {
		name          string
		pvzID         string
		mockSetup     func(*mockProductStorage)
		expectedError string
	}{
		{
			name:  "Success",
			pvzID: "pvz1",
			mockSetup: func(ps *mockProductStorage) {
				ps.On("DeleteLastProduct", "pvz1").Return(nil)
			},
			expectedError: "",
		},
		{
			name:          "Empty_PVZID",
			pvzID:         "",
			mockSetup:     func(ps *mockProductStorage) {},
			expectedError: "pvzID is required",
		},
		{
			name:  "Storage_Error",
			pvzID: "pvz1",
			mockSetup: func(ps *mockProductStorage) {
				ps.On("DeleteLastProduct", "pvz1").Return(errors.New("storage error"))
			},
			expectedError: "storage error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка мока
			productStorage := &mockProductStorage{}
			receptionStorage := &mockReceptionStorage{}
			tt.mockSetup(productStorage)

			// Создание сервиса
			service := NewProductService(productStorage, receptionStorage)

			// Выполнение метода
			err := service.DeleteLastProduct(tt.pvzID)

			// Проверка результатов
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}

			// Проверка вызовов мока
			productStorage.AssertExpectations(t)
			receptionStorage.AssertExpectations(t)
		})
	}
}
