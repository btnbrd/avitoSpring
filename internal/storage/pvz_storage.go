package storage

import "avitoSpring/internal/models"

type PVZStorageI interface {
	CreatePVZ(pvz *models.PVZ) (string, error)

	GetPVZs(filterDate string, page, pageSize int) ([]*models.PVZ, error)
}
