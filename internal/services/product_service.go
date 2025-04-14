package services

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/storage"
	"fmt"
	"time"
)

type ProductService struct {
	productStorage   storage.ProductStorage
	receptionStorage storage.ReceptionStorage
}

type ProductServiceInterface interface {
	CreateProduct(product *models.Product, pvzID string) (string, error)
	DeleteLastProduct(pvzID string) error
}

var _ ProductServiceInterface = (*ProductService)(nil)

func NewProductService(store storage.ProductStorage, receptionStorage storage.ReceptionStorage) *ProductService {
	return &ProductService{productStorage: store, receptionStorage: receptionStorage}
}

func (s *ProductService) CreateProduct(product *models.Product, pvzID string) (string, error) {

	if product.Type != models.ProductTypeElectronics && product.Type != models.ProductTypeClothing && product.Type != models.ProductTypeFootwear {
		return "", fmt.Errorf("invalid product type: %s", product.Type)
	}

	if product.DateTime == "" {
		product.DateTime = time.Now().Format(time.RFC3339)
	}

	reception, err := s.receptionStorage.GetLastReceptionByPVZID(pvzID)
	if err != nil {
		return "", fmt.Errorf("failed to get last reception: %w", err)
	}
	if reception == nil || reception.Status != models.ReceptionStatusInProgress {
		return "", fmt.Errorf("no open reception found for PVZ %s", pvzID)
	}

	product.ReceptionID = reception.ID

	return s.productStorage.CreateProduct(product)
}

func (s *ProductService) DeleteLastProduct(pvzID string) error {
	if pvzID == "" {
		return fmt.Errorf("pvzID is required")
	}

	return s.productStorage.DeleteLastProduct(pvzID)
}
