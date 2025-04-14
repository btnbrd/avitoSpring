package storage

import "avitoSpring/internal/models"

type ReceptionStorage interface {
	CreateReception(reception *models.Reception) (string, error)
	HasOpenReception(pvzID string) (bool, error)
	CloseLastReception(pvzID string) error
	GetLastReceptionByPVZID(pvzID string) (*models.Reception, error)
}
