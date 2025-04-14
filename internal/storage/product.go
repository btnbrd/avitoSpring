package storage

import "avitoSpring/internal/models"

type ProductStorage interface {
	CreateProduct(product *models.Product) (string, error)
	DeleteLastProduct(pvzID string) error
}
