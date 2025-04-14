package services

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/storage"
	"fmt"
	"time"
)

type PVZService struct {
	store storage.PVZStorageI
}

func NewPVZService(store storage.PVZStorageI) *PVZService {
	return &PVZService{store: store}
}

func (s *PVZService) CreatePVZ(pvz *models.PVZ) (string, error) {
	if pvz.City != models.CityMoscow && pvz.City != models.CitySaintPetersburg && pvz.City != models.CityKazan {
		return "", fmt.Errorf("invalid city: %s", pvz.City)
	}

	if pvz.RegistrationDate == "" {
		pvz.RegistrationDate = time.Now().Format(time.RFC3339)
	}

	return s.store.CreatePVZ(pvz)
}

func (s *PVZService) GetPVZs(filterDate string, page, pageSize int) ([]*models.PVZ, error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be greater than 0")
	}
	if pageSize < 1 {
		return nil, fmt.Errorf("pageSize must be greater than 0")
	}

	if filterDate != "" {
	}

	return s.store.GetPVZs(filterDate, page, pageSize)
}

func (s *PVZService) GetPVZsWithDetails(startDate, endDate string, page, pageSize int) ([]*models.PVZWithDetails, error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be greater than 0")
	}
	if pageSize < 1 {
		return nil, fmt.Errorf("pageSize must be greater than 0")
	}

	return s.store.GetPVZsWithDetails(startDate, endDate, page, pageSize)
}
