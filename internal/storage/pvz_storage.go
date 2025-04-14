package storage

import "avitoSpring/internal/models"

type PVZStorageI interface {
	CreatePVZ(pvz *models.PVZ) (string, error)
	GetPVZs(filterDate string, page, pageSize int) ([]*models.PVZ, error)
	GetPVZsWithDetails(startDate, endDate string, page, pageSize int) ([]*models.PVZWithDetails, error)
	//CloseLastReception(pvzID string) error
	//GetLastReceptionByPVZID(pvzID string) (*models.Reception, error)
	//DeleteLastProduct(pvzID string) error
}
