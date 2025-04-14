package services

import (
	"avitoSpring/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// Mock реализации интерфейса PVZStorageI
type MockPVZStorage struct {
	mock.Mock
}

func (m *MockPVZStorage) CreatePVZ(pvz *models.PVZ) (string, error) {
	args := m.Called(pvz)
	return args.String(0), args.Error(1)
}

func (m *MockPVZStorage) GetPVZsWithDetails(startDate, endDate string, page, pageSize int) ([]*models.PVZWithDetails, error) {
	args := m.Called(startDate, endDate, page, pageSize)
	return args.Get(0).([]*models.PVZWithDetails), args.Error(1)
}

func TestCreatePVZ_ValidCity(t *testing.T) {
	mockStore := new(MockPVZStorage)
	service := NewPVZService(mockStore)

	pvz := &models.PVZ{
		City: models.CityMoscow,
	}
	mockStore.On("CreatePVZ", mock.Anything).Return("123", nil)

	id, err := service.CreatePVZ(pvz)

	assert.NoError(t, err)
	assert.Equal(t, "123", id)
	assert.NotEmpty(t, pvz.RegistrationDate)
}

func TestCreatePVZ_InvalidCity(t *testing.T) {
	mockStore := new(MockPVZStorage)
	service := NewPVZService(mockStore)

	pvz := &models.PVZ{
		City: "Новосибирск",
	}

	id, err := service.CreatePVZ(pvz)

	assert.Error(t, err)
	assert.Equal(t, "", id)
	assert.Contains(t, err.Error(), "invalid city")
}

func TestGetPVZsWithDetails_ValidRequest(t *testing.T) {
	mockStore := new(MockPVZStorage)
	service := NewPVZService(mockStore)

	expected := []*models.PVZWithDetails{}
	mockStore.On("GetPVZsWithDetails", "", "", 1, 10).Return(expected, nil)

	res, err := service.GetPVZsWithDetails("", "", 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestGetPVZsWithDetails_InvalidPage(t *testing.T) {
	mockStore := new(MockPVZStorage)
	service := NewPVZService(mockStore)

	res, err := service.GetPVZsWithDetails("", "", 0, 10)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "page must be greater than 0", err.Error())
}

func TestGetPVZsWithDetails_InvalidPageSize(t *testing.T) {
	mockStore := new(MockPVZStorage)
	service := NewPVZService(mockStore)

	res, err := service.GetPVZsWithDetails("", "", 1, 0)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "pageSize must be greater than 0", err.Error())
}
