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
        INSERT INTO pvzs (id, registration_date, city)
        VALUES ($1, $2, $3)
        RETURNING id`
	err := p.DB.QueryRow(query, pvzID, pvz.RegistrationDate, pvz.City).Scan(&pvzID)
	if err != nil {
		return "", fmt.Errorf("failed to create PVZ: %w", err)
	}

	return pvzID, nil
}

func (p PVZStoragePG) GetPVZs(filterDate string, page, pageSize int) ([]*models.PVZ, error) {
	var pvzs []*models.PVZ
	offset := (page - 1) * pageSize

	query := `
        SELECT id, registration_date, city
        FROM pvzs
        WHERE ($1 = '' OR registration_date >= $1)
        ORDER BY registration_date
        LIMIT $2 OFFSET $3`
	rows, err := p.DB.Query(query, filterDate, pageSize, offset)
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
		pvzs = append(pvzs, &pvz)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating PVZs: %w", err)
	}

	return pvzs, nil
}

var _ storage.PVZStorageI = (*PVZStoragePG)(nil)

func NewPVZStorage(db *sql.DB) *PVZStoragePG {

	return &PVZStoragePG{DB: db}
}
