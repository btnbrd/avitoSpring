package pg

import (
	"avitoSpring/internal/models"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPVZStoragePG_CreatePVZ(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := NewPVZStorage(db)

	tests := []struct {
		name          string
		pvz           *models.PVZ
		mockSetup     func()
		expectedID    string
		expectedError string
	}{
		{
			name: "Success",
			pvz: &models.PVZ{
				RegistrationDate: "2023-01-01",
				City:             "Moscow",
			},
			mockSetup: func() {
				pvzID := uuid.New().String()
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3) RETURNING id`)).
					WithArgs(sqlmock.AnyArg(), "2023-01-01", "Moscow").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pvzID))
			},
			expectedID:    "",
			expectedError: "",
		},
		{
			name: "DB_Error",
			pvz: &models.PVZ{
				RegistrationDate: "2023-01-01",
				City:             "Moscow",
			},
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3) RETURNING id`)).
					WithArgs(sqlmock.AnyArg(), "2023-01-01", "Moscow").
					WillReturnError(errors.New("db error"))
			},
			expectedID:    "",
			expectedError: "failed to create PVZ: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			id, err := storage.CreatePVZ(tt.pvz)

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

//func TestPVZStoragePG_GetPVZs(t *testing.T) {
//	db, mock, err := sqlmock.New()
//	assert.NoError(t, err)
//	defer db.Close()
//
//	storage := NewPVZStorage(db)
//
//	tests := []struct {
//		name          string
//		filterDate    string
//		page          int
//		pageSize      int
//		mockSetup     func()
//		expectedPVZs  []*models.PVZ
//		expectedError string
//	}{
//		{
//			name:       "Success_With_Filter",
//			filterDate: "2023-01-01",
//			page:       1,
//			pageSize:   2,
//			mockSetup: func() {
//				mock.ExpectQuery(regexp.QuoteMeta(
//					`SELECT id, registration_date, city FROM pvz WHERE ($1 = '' OR registration_date >= $1) ORDER BY registration_date LIMIT $2 OFFSET $3`)).
//					WithArgs("2023-01-01", 2, 0).
//					WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
//						AddRow("pvz1", "2023-01-01", "Moscow").
//						AddRow("pvz2", "2023-01-02", "Berlin"))
//			},
//			expectedPVZs: []*models.PVZ{
//				{ID: "pvz1", RegistrationDate: "2023-01-01", City: "Moscow"},
//				{ID: "pvz2", RegistrationDate: "2023-01-02", City: "Berlin"},
//			},
//			expectedError: "",
//		},
//		{
//			name:       "Success_No_Filter",
//			filterDate: "",
//			page:       2,
//			pageSize:   1,
//			mockSetup: func() {
//				mock.ExpectQuery(regexp.QuoteMeta(
//					`SELECT id, registration_date, city FROM pvz WHERE ($1 = '' OR registration_date >= $1) ORDER BY registration_date LIMIT $2 OFFSET $3`)).
//					WithArgs("", 1, 1).
//					WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
//						AddRow("pvz3", "2023-01-03", "Paris"))
//			},
//			expectedPVZs: []*models.PVZ{
//				{ID: "pvz3", RegistrationDate: "2023-01-03", City: "Paris"},
//			},
//			expectedError: "",
//		},
//		{
//			name:       "No_Results",
//			filterDate: "2023-01-01",
//			page:       1,
//			pageSize:   2,
//			mockSetup: func() {
//				mock.ExpectQuery(regexp.QuoteMeta(
//					`SELECT id, registration_date, city FROM pvz WHERE ($1 = '' OR registration_date >= $1) ORDER BY registration_date LIMIT $2 OFFSET $3`)).
//					WithArgs("2023-01-01", 2, 0).
//					WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}))
//			},
//			expectedPVZs:  []*models.PVZ{nil},
//			expectedError: "",
//		},
//		{
//			name:       "DB_Error",
//			filterDate: "2023-01-01",
//			page:       1,
//			pageSize:   2,
//			mockSetup: func() {
//				mock.ExpectQuery(regexp.QuoteMeta(
//					`SELECT id, registration_date, city FROM pvz WHERE ($1 = '' OR registration_date >= $1) ORDER BY registration_date LIMIT $2 OFFSET $3`)).
//					WithArgs("2023-01-01", 2, 0).
//					WillReturnError(errors.New("db error"))
//			},
//			expectedPVZs:  nil,
//			expectedError: "failed to get PVZs: db error",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			tt.mockSetup()
//
//			pvzs, err := storage.GetPVZs(tt.filterDate, tt.page, tt.pageSize)
//
//			assert.Equal(t, tt.expectedPVZs, pvzs)
//			if tt.expectedError == "" {
//				assert.NoError(t, err)
//			} else {
//				assert.EqualError(t, err, tt.expectedError)
//			}
//
//			assert.NoError(t, mock.ExpectationsWereMet())
//		})
//	}
//}

func TestPVZStoragePG_GetPVZsWithDetails(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := NewPVZStorage(db)

	tests := []struct {
		name          string
		startDate     string
		endDate       string
		page          int
		pageSize      int
		mockSetup     func()
		expectedPVZs  []*models.PVZWithDetails
		expectedError string
	}{
		{
			name:      "Success_With_Dates",
			startDate: "2023-01-01",
			endDate:   "2023-01-02",
			page:      1,
			pageSize:  1,
			mockSetup: func() {
				// PVZ query
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, registration_date, city FROM pvz ORDER BY registration_date LIMIT $1 OFFSET $2`)).
					WithArgs(1, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
						AddRow("pvz1", "2023-01-01", "Moscow"))

				// Receptions query
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, datetime, pvz_id, status FROM receptions WHERE pvz_id = $1 AND datetime BETWEEN $2 AND $3 ORDER BY datetime`)).
					WithArgs("pvz1", "2023-01-01", "2023-01-02").
					WillReturnRows(sqlmock.NewRows([]string{"id", "datetime", "pvz_id", "status"}).
						AddRow("rec1", "2023-01-01T12:00:00Z", "pvz1", "in_progress"))

				// Products query
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, datetime, type, reception_id FROM products WHERE reception_id = $1 ORDER BY datetime`)).
					WithArgs("rec1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "datetime", "type", "reception_id"}).
						AddRow("prod1", "2023-01-01T12:01:00Z", "electronics", "rec1"))
			},
			expectedPVZs: []*models.PVZWithDetails{
				{
					PVZ: &models.PVZ{
						ID:               "pvz1",
						RegistrationDate: "2023-01-01",
						City:             "Moscow",
					},
					Receptions: []*models.ReceptionWithProducts{
						{
							Reception: &models.Reception{
								ID:       "rec1",
								DateTime: "2023-01-01T12:00:00Z",
								PVZID:    "pvz1",
								Status:   "in_progress",
							},
							Products: []*models.Product{
								{
									ID:          "prod1",
									DateTime:    "2023-01-01T12:01:00Z",
									Type:        "electronics",
									ReceptionID: "rec1",
								},
							},
						},
					},
				},
			},
			expectedError: "",
		},
		{
			name:      "No_Results",
			startDate: "",
			endDate:   "",
			page:      1,
			pageSize:  1,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, registration_date, city FROM pvz ORDER BY registration_date LIMIT $1 OFFSET $2`)).
					WithArgs(1, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}))
			},
			expectedPVZs:  []*models.PVZWithDetails{},
			expectedError: "",
		},
		{
			name:      "DB_Error_PVZ",
			startDate: "",
			endDate:   "",
			page:      1,
			pageSize:  1,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, registration_date, city FROM pvz ORDER BY registration_date LIMIT $1 OFFSET $2`)).
					WithArgs(1, 0).
					WillReturnError(errors.New("db error"))
			},
			expectedPVZs:  nil,
			expectedError: "failed to get PVZs: db error",
		},
		{
			name:      "DB_Error_Receptions",
			startDate: "",
			endDate:   "",
			page:      1,
			pageSize:  1,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, registration_date, city FROM pvz ORDER BY registration_date LIMIT $1 OFFSET $2`)).
					WithArgs(1, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
						AddRow("pvz1", "2023-01-01", "Moscow"))
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, datetime, pvz_id, status FROM receptions WHERE pvz_id = $1 ORDER BY datetime`)).
					WithArgs("pvz1").
					WillReturnError(errors.New("db error"))
			},
			expectedPVZs:  nil,
			expectedError: "failed to get receptions for PVZ pvz1: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			pvzs, err := storage.GetPVZsWithDetails(tt.startDate, tt.endDate, tt.page, tt.pageSize)

			assert.Equal(t, tt.expectedPVZs, pvzs)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
