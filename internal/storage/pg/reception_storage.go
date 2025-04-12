package pg

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/storage"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
)

type ReceptionStoragePG struct {
	DB *sql.DB
}

func (r ReceptionStoragePG) CreateReception(reception *models.Reception) (string, error) {
	receptionID := uuid.New().String()

	query := `
        INSERT INTO receptions (id, date_time, pvz_id, status)
        VALUES ($1, $2, $3, $4)
        RETURNING id`
	err := r.DB.QueryRow(query, receptionID, reception.DateTime, reception.PVZID, reception.Status).Scan(&receptionID)
	if err != nil {
		return "", fmt.Errorf("failed to create reception: %w", err)
	}

	return receptionID, nil
}

var _ storage.ReceptionStorage = (*ReceptionStoragePG)(nil)

func NewReceptionStorage(db *sql.DB) *ReceptionStoragePG {

	return &ReceptionStoragePG{DB: db}
}
