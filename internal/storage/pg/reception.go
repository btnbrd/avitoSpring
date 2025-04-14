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

func (r *ReceptionStoragePG) CloseLastReception(pvzID string) error {
	var receptionID string
	query := `
        SELECT id
        FROM receptions
        WHERE pvz_id = $1 AND status = 'in_progress'
        ORDER BY datetime DESC
        LIMIT 1`
	err := r.DB.QueryRow(query, pvzID).Scan(&receptionID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no open reception found for PVZ %s", pvzID)
	}
	if err != nil {
		return fmt.Errorf("failed to find open reception: %w", err)
	}

	updateQuery := `
        UPDATE receptions
        SET status = 'close'
        WHERE id = $1`
	_, err = r.DB.Exec(updateQuery, receptionID)
	if err != nil {
		return fmt.Errorf("failed to close reception: %w", err)
	}

	return nil
}

func (r *ReceptionStoragePG) HasOpenReception(pvzID string) (bool, error) {
	var count int
	query := `
        SELECT COUNT(*)
        FROM receptions
        WHERE pvz_id = $1 AND status = 'in_progress'`
	err := r.DB.QueryRow(query, pvzID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check open receptions: %w", err)
	}
	return count > 0, nil
}

func (r *ReceptionStoragePG) CreateReception(reception *models.Reception) (string, error) {
	receptionID := uuid.New().String()

	query := `
        INSERT INTO receptions (id, datetime, pvz_id, status)
        VALUES ($1, $2, $3, $4)
        RETURNING id`
	err := r.DB.QueryRow(query, receptionID, reception.DateTime, reception.PVZID, reception.Status).Scan(&receptionID)
	if err != nil {
		return "", fmt.Errorf("failed to create reception: %w", err)
	}

	return receptionID, nil
}

func (p *ReceptionStoragePG) GetLastReceptionByPVZID(pvzID string) (*models.Reception, error) {
	var reception models.Reception
	query := `
        SELECT id, datetime, pvz_id, status
        FROM receptions
        WHERE pvz_id = $1
        ORDER BY datetime DESC
        LIMIT 1`
	err := p.DB.QueryRow(query, pvzID).Scan(&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last reception: %w", err)
	}
	return &reception, nil
}

var _ storage.ReceptionStorage = (*ReceptionStoragePG)(nil)

func NewReceptionStorage(db *sql.DB) *ReceptionStoragePG {

	return &ReceptionStoragePG{DB: db}
}
