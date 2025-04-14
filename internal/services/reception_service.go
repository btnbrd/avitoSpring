package services

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/storage"
	"fmt"
	"time"
)

type ReceptionServiceInterface interface {
	CreateReception(reception *models.Reception) (string, error)
	CloseLastReception(pvzID string) error
	GetLastReceptionByPVZID(pvzID string) (*models.Reception, error)
}

var _ ReceptionServiceInterface = (*ReceptionService)(nil)

type ReceptionService struct {
	store storage.ReceptionStorage
}

func NewReceptionService(store storage.ReceptionStorage) *ReceptionService {
	return &ReceptionService{store: store}
}

func (s *ReceptionService) CreateReception(reception *models.Reception) (string, error) {
	if reception.PVZID == "" {
		return "", fmt.Errorf("pvzId is required")
	}
	hasOpen, err := s.store.HasOpenReception(reception.PVZID)
	if err != nil {
		return "", fmt.Errorf("failed to check open receptions: %w", err)
	}
	if hasOpen {
		return "", fmt.Errorf("an open reception already exists for PVZ %s", reception.PVZID)
	}

	if reception.DateTime == "" {
		reception.DateTime = time.Now().Format(time.RFC3339)
	}

	if reception.Status == "" {
		reception.Status = models.ReceptionStatusInProgress
	}

	if reception.Status != models.ReceptionStatusInProgress && reception.Status != models.ReceptionStatusClose {
		return "", fmt.Errorf("invalid status: %s", reception.Status)
	}

	return s.store.CreateReception(reception)
}

func (s *ReceptionService) GetLastReceptionByPVZID(pvzID string) (*models.Reception, error) {
	if pvzID == "" {
		return nil, fmt.Errorf("pvzID is required")
	}

	return s.store.GetLastReceptionByPVZID(pvzID)
}

func (s *ReceptionService) CloseLastReception(pvzID string) error {
	if pvzID == "" {
		return fmt.Errorf("pvzID is required")
	}

	return s.store.CloseLastReception(pvzID)
}
