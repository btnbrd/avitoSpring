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

func TestProductStoragePG_CreateProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := NewProductStorage(db)

	tests := []struct {
		name          string
		product       *models.Product
		mockSetup     func()
		expectedID    string
		expectedError string
	}{
		{
			name: "Success",
			product: &models.Product{
				DateTime:    "2023-01-01T12:00:00Z",
				Type:        models.ProductTypeElectronics,
				ReceptionID: "rec1",
			},
			mockSetup: func() {
				productID := uuid.New().String()
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO products (id, datetime, type, reception_id) VALUES ($1, $2, $3, $4) RETURNING id`)).
					WithArgs(sqlmock.AnyArg(), "2023-01-01T12:00:00Z", models.ProductTypeElectronics, "rec1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(productID))
			},
			expectedID:    "",
			expectedError: "",
		},
		{
			name: "DB_Error",
			product: &models.Product{
				DateTime:    "2023-01-01T12:00:00Z",
				Type:        models.ProductTypeElectronics,
				ReceptionID: "rec1",
			},
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO products (id, datetime, type, reception_id) VALUES ($1, $2, $3, $4) RETURNING id`)).
					WithArgs(sqlmock.AnyArg(), "2023-01-01T12:00:00Z", models.ProductTypeElectronics, "rec1").
					WillReturnError(errors.New("db error"))
			},
			expectedID:    "",
			expectedError: "failed to create product: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			id, err := storage.CreateProduct(tt.product)

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

func TestProductStoragePG_DeleteLastProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := NewProductStorage(db)

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

				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM products WHERE reception_id = $1 ORDER BY datetime DESC LIMIT 1`)).
					WithArgs("rec1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("prod1"))

				mock.ExpectExec(regexp.QuoteMeta(
					`DELETE FROM products WHERE id = $1`)).
					WithArgs("prod1").
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
			name:  "No_Products",
			pvzID: "pvz1",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM receptions WHERE pvz_id = $1 AND status = 'in_progress' ORDER BY datetime DESC LIMIT 1`)).
					WithArgs("pvz1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("rec1"))

				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM products WHERE reception_id = $1 ORDER BY datetime DESC LIMIT 1`)).
					WithArgs("rec1").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: "no products found in reception rec1",
		},
		{
			name:  "DB_Error_Reception",
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
			name:  "DB_Error_Delete",
			pvzID: "pvz1",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM receptions WHERE pvz_id = $1 AND status = 'in_progress' ORDER BY datetime DESC LIMIT 1`)).
					WithArgs("pvz1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("rec1"))

				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM products WHERE reception_id = $1 ORDER BY datetime DESC LIMIT 1`)).
					WithArgs("rec1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("prod1"))

				mock.ExpectExec(regexp.QuoteMeta(
					`DELETE FROM products WHERE id = $1`)).
					WithArgs("prod1").
					WillReturnError(errors.New("db error"))
			},
			expectedError: "failed to delete product: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := storage.DeleteLastProduct(tt.pvzID)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
