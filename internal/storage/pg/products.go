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
        INSERT INTO products (id, datetime, type, reception_id)
        VALUES ($1, $2, $3, $4)
        RETURNING id`
	err := s.DB.QueryRow(query, productID, product.DateTime, product.Type, product.ReceptionID).Scan(&productID)
	if err != nil {
		return "", fmt.Errorf("failed to create product: %w", err)
	}

	return productID, nil
}

func (p *ProductStoragePG) DeleteLastProduct(pvzID string) error {
	var receptionID string
	query := `
        SELECT id
        FROM receptions
        WHERE pvz_id = $1 AND status = 'in_progress'
        ORDER BY datetime DESC
        LIMIT 1`
	err := p.DB.QueryRow(query, pvzID).Scan(&receptionID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no open reception found for PVZ %s", pvzID)
	}
	if err != nil {
		return fmt.Errorf("failed to find open reception: %w", err)
	}

	var productID string
	productQuery := `
        SELECT id
        FROM products
        WHERE reception_id = $1
        ORDER BY datetime DESC
        LIMIT 1`
	err = p.DB.QueryRow(productQuery, receptionID).Scan(&productID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no products found in reception %s", receptionID)
	}
	if err != nil {
		return fmt.Errorf("failed to find last product: %w", err)
	}

	deleteQuery := `
        DELETE FROM products
        WHERE id = $1`
	_, err = p.DB.Exec(deleteQuery, productID)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func NewProductStorage(db *sql.DB) *ProductStoragePG {

	return &ProductStoragePG{DB: db}
}
