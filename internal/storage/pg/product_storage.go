package pg

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/storage"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
)

type ProductStoragePG struct {
	DB *sql.DB
}

var _ storage.ProductStorage = (*ProductStoragePG)(nil)

func (s *ProductStoragePG) CreateProduct(product *models.Product) (string, error) {
	productID := uuid.New().String()

	query := `
        INSERT INTO products (id, date_time, type, reception_id)
        VALUES ($1, $2, $3, $4)
        RETURNING id`
	err := s.DB.QueryRow(query, productID, product.DateTime, product.Type, product.ReceptionID).Scan(&productID)
	if err != nil {
		return "", fmt.Errorf("failed to create product: %w", err)
	}

	return productID, nil
}

func NewProductStorage(db *sql.DB) *ProductStoragePG {

	return &ProductStoragePG{DB: db}
}
