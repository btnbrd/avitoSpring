package pg

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/storage"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
)

type PVZStoragePG struct {
	DB *sql.DB
}

func (p PVZStoragePG) CreatePVZ(pvz *models.PVZ) (string, error) {
	//TODO implement me
	pvzID := uuid.New().String()

	query := `
        INSERT INTO pvz (id, registration_date, city)
        VALUES ($1, $2, $3)
        RETURNING id`
	err := p.DB.QueryRow(query, pvzID, pvz.RegistrationDate, pvz.City).Scan(&pvzID)
	if err != nil {
		return "", fmt.Errorf("failed to create PVZ: %w", err)
	}

	return pvzID, nil
}

var _ storage.PVZStorageI = (*PVZStoragePG)(nil)

func NewPVZStorage(db *sql.DB) *PVZStoragePG {

	return &PVZStoragePG{DB: db}
}

func (s *PVZStoragePG) GetPVZsWithDetails(startDate, endDate string, page, pageSize int) ([]*models.PVZWithDetails, error) {
	var pvzDetails []*models.PVZWithDetails
	pvzDetails = make([]*models.PVZWithDetails, 0)
	offset := (page - 1) * pageSize

	query := `
        SELECT id, registration_date, city
        FROM pvz
        ORDER BY registration_date
        LIMIT $1 OFFSET $2`
	rows, err := s.DB.Query(query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get PVZs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pvz models.PVZ
		err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
		if err != nil {
			return nil, fmt.Errorf("failed to scan PVZ: %w", err)
		}

		var receptions []*models.ReceptionWithProducts
		receptionQuery := `
            SELECT id, datetime, pvz_id, status
            FROM receptions
            WHERE pvz_id = $1`
		params := []interface{}{pvz.ID}

		if startDate != "" && endDate != "" {
			receptionQuery += ` AND datetime BETWEEN $2 AND $3`
			params = append(params, startDate, endDate)
		} else if startDate != "" {
			receptionQuery += ` AND datetime >= $2`
			params = append(params, startDate)
		} else if endDate != "" {
			receptionQuery += ` AND datetime <= $2`
			params = append(params, endDate)
		}

		receptionQuery += "\n ORDER BY datetime;"

		receptionRows, err := s.DB.Query(receptionQuery, params...)
		if err != nil {
			return nil, fmt.Errorf("failed to get receptions for PVZ %s: %w", pvz.ID, err)
		}
		defer receptionRows.Close()

		for receptionRows.Next() {
			var reception models.Reception
			err := receptionRows.Scan(&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status)
			if err != nil {
				return nil, fmt.Errorf("failed to scan reception: %w", err)
			}

			var products []*models.Product
			productQuery := `
                SELECT id, datetime, type, reception_id
                FROM products
                WHERE reception_id = $1
                ORDER BY datetime`
			productRows, err := s.DB.Query(productQuery, reception.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get products for reception %s: %w", reception.ID, err)
			}
			defer productRows.Close()

			for productRows.Next() {
				var product models.Product
				err := productRows.Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionID)
				if err != nil {
					return nil, fmt.Errorf("failed to scan product: %w", err)
				}
				products = append(products, &product)
			}

			if err := productRows.Err(); err != nil {
				return nil, fmt.Errorf("error iterating products: %w", err)
			}

			receptions = append(receptions, &models.ReceptionWithProducts{
				Reception: &reception,
				Products:  products,
			})
		}

		if err := receptionRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating receptions: %w", err)
		}

		pvzDetails = append(pvzDetails, &models.PVZWithDetails{
			PVZ:        &pvz,
			Receptions: receptions,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating PVZs: %w", err)
	}

	return pvzDetails, nil
}
