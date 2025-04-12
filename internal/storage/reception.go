package storage

import "avitoSpring/internal/models"

type ReceptionStorage interface {
	CreateReception(reception *models.Reception) (string, error)
}
