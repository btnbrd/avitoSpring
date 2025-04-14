package pg

import (
	"avitoSpring/internal/models"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestReceptionStoragePG_CreateReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := NewReceptionStorage(db)

	tests := []struct {
		name          string
		reception     *models.Reception
		mockSetup     func()
		expectedID    string
		expectedError string
	}{
		{
			name: "Success",
			reception: &models.Reception{
				DateTime: "2023-01-01T12:00:00Z",
				PVZID:    "pvz1",
				Status:   "in_progress",
			},
			mockSetup: func() {
				receptionID := uuid.New().String()
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO receptions (id, datetime, pvz_id, status) VALUES ($1, $2, $3, $4) RETURNING id`)).
					WithArgs(sqlmock.AnyArg(), "2023-01-01T12:00:00Z", "pvz1", "in_progress").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(receptionID))
			},
			expectedID:    "",
			expectedError: "",
		},
		{
			name: "DB_Error",
			reception: &models.Reception{
				DateTime: "2023-01-01T12:00:00Z",
				PVZID:    "pvz1",
				Status:   "in_progress",
			},
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO receptions (id, datetime, pvz_id, status) VALUES ($1, $2, $3, $4) RETURNING id`)).
					WithArgs(sqlmock.AnyArg(), "2023-01-01T12:00:00Z", "pvz1", "in_progress").
					WillReturnError(errors.New("db error"))
			},
			expectedID:    "",
			expectedError: "failed to create reception: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			id, err := storage.CreateReception(tt.reception)

			if tt.expectedError == "" {
				assert.NoError(t, err)
				assert.NotEmpty(t, id)
				_, parseErr := uuid.Parse(id)
				assert.NoError(t, parseErr)
			} else {
				assert.EqualError(t, err, tt.expectedError)
				assert.Empty(t, id)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionStoragePG_HasOpenReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := NewReceptionStorage(db)

	tests := []struct {
		name           string
		pvzID          string
		mockSetup      func()
		expectedResult bool
		expectedError  string
	}{
		{
			name:  "Has_Open_Reception",
			pvzID: "pvz1",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(*) FROM receptions WHERE pvz_id = $1 AND status = 'in_progress'`)).
					WithArgs("pvz1").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedResult: true,
			expectedError:  "",
		},
		{
			name:  "No_Open_Reception",
			pvzID: "pvz1",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(*) FROM receptions WHERE pvz_id = $1 AND status = 'in_progress'`)).
					WithArgs("pvz1").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedResult: false,
			expectedError:  "",
		},
		{
			name:  "DB_Error",
			pvzID: "pvz1",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(*) FROM receptions WHERE pvz_id = $1 AND status = 'in_progress'`)).
					WithArgs("pvz1").
					WillReturnError(errors.New("db error"))
			},
			expectedResult: false,
			expectedError:  "failed to check open receptions: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := storage.HasOpenReception(tt.pvzID)

			assert.Equal(t, tt.expectedResult, result)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionStoragePG_CloseLastReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := NewReceptionStorage(db)

	tests := []struct {
		name          string
		pvzID         string
		mockSetup     func()
		expectedError string
	}{
		{
			name:  "Success",
			pvzID: "pvz1",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM receptions WHERE pvz_id = $1 AND status = 'in_progress' ORDER BY datetime DESC LIMIT 1`)).
					WithArgs("pvz1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("rec1"))

				mock.ExpectExec(regexp.QuoteMeta(
					`UPDATE receptions SET status = 'close' WHERE id = $1`)).
					WithArgs("rec1").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: "",
		},
		{
			name:  "No_Open_Reception",
			pvzID: "pvz1",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM receptions WHERE pvz_id = $1 AND status = 'in_progress' ORDER BY datetime DESC LIMIT 1`)).
					WithArgs("pvz1").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: "no open reception found for PVZ pvz1",
		},
		{
			name:  "DB_Error_Find",
			pvzID: "pvz1",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM receptions WHERE pvz_id = $1 AND status = 'in_progress' ORDER BY datetime DESC LIMIT 1`)).
					WithArgs("pvz1").
					WillReturnError(errors.New("db error"))
			},
			expectedError: "failed to find open reception: db error",
		},
		{
			name:  "DB_Error_Update",
			pvzID: "pvz1",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM receptions WHERE pvz_id = $1 AND status = 'in_progress' ORDER BY datetime DESC LIMIT 1`)).
					WithArgs("pvz1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("rec1"))

				mock.ExpectExec(regexp.QuoteMeta(
					`UPDATE receptions SET status = 'close' WHERE id = $1`)).
					WithArgs("rec1").
					WillReturnError(errors.New("db error"))
			},
			expectedError: "failed to close reception: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := storage.CloseLastReception(tt.pvzID)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReceptionStoragePG_GetLastReceptionByPVZID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := NewReceptionStorage(db)

	tests := []struct {
		name              string
		pvzID             string
		mockSetup         func()
		expectedReception *models.Reception
		expectedError     string
	}{
		{
			name:  "Success",
			pvzID: "pvz1",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, datetime, pvz_id, status FROM receptions WHERE pvz_id = $1 ORDER BY datetime DESC LIMIT 1`)).
					WithArgs("pvz1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "datetime", "pvz_id", "status"}).
						AddRow("rec1", "2023-01-01T12:00:00Z", "pvz1", "in_progress"))
			},
			expectedReception: &models.Reception{
				ID:       "rec1",
				DateTime: "2023-01-01T12:00:00Z",
				PVZID:    "pvz1",
				Status:   "in_progress",
			},
			expectedError: "",
		},
		{
			name:  "No_Reception",
			pvzID: "pvz1",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, datetime, pvz_id, status FROM receptions WHERE pvz_id = $1 ORDER BY datetime DESC LIMIT 1`)).
					WithArgs("pvz1").
					WillReturnError(sql.ErrNoRows)
			},
			expectedReception: nil,
			expectedError:     "",
		},
		{
			name:  "DB_Error",
			pvzID: "pvz1",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, datetime, pvz_id, status FROM receptions WHERE pvz_id = $1 ORDER BY datetime DESC LIMIT 1`)).
					WithArgs("pvz1").
					WillReturnError(errors.New("db error"))
			},
			expectedReception: nil,
			expectedError:     "failed to get last reception: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			reception, err := storage.GetLastReceptionByPVZID(tt.pvzID)

			assert.Equal(t, tt.expectedReception, reception)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
