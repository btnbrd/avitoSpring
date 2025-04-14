package pg

import (
	"avitoSpring/internal/models"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUserStoragePG_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := NewUserStorage(db)

	tests := []struct {
		name          string
		user          *models.User
		password      string
		mockSetup     func()
		expectedID    string
		expectedError string
	}{
		{
			name: "Success",
			user: &models.User{
				Email: "test@example.com",
				Role:  "moderator",
			},
			password: "password123",
			mockSetup: func() {
				userID := uuid.New().String()
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO users (id, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id`)).
					WithArgs(sqlmock.AnyArg(), "test@example.com", sqlmock.AnyArg(), "moderator").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))
			},
			expectedID:    "",
			expectedError: "",
		},
		{
			name: "DB_Error",
			user: &models.User{
				Email: "test@example.com",
				Role:  "moderator",
			},
			password: "password123",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO users (id, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id`)).
					WithArgs(sqlmock.AnyArg(), "test@example.com", sqlmock.AnyArg(), "moderator").
					WillReturnError(errors.New("db error"))
			},
			expectedID:    "",
			expectedError: "failed to create user: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			id, err := storage.CreateUser(tt.user, tt.password)

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

func TestUserStoragePG_GetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := NewUserStorage(db)

	tests := []struct {
		name          string
		email         string
		mockSetup     func()
		expectedUser  *models.User
		expectedError string
	}{
		{
			name:  "Success",
			email: "test@example.com",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, email, role FROM users WHERE email = $1`)).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "role"}).
						AddRow("user1", "test@example.com", "moderator"))
			},
			expectedUser: &models.User{
				ID:    "user1",
				Email: "test@example.com",
				Role:  "moderator",
			},
			expectedError: "",
		},
		{
			name:  "Not_Found",
			email: "notfound@example.com",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, email, role FROM users WHERE email = $1`)).
					WithArgs("notfound@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: "",
		},
		{
			name:  "DB_Error",
			email: "test@example.com",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, email, role FROM users WHERE email = $1`)).
					WithArgs("test@example.com").
					WillReturnError(errors.New("db error"))
			},
			expectedUser:  nil,
			expectedError: "failed to get user by email: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			user, err := storage.GetUserByEmail(tt.email)

			assert.Equal(t, tt.expectedUser, user)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserStoragePG_CheckPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := NewUserStorage(db)

	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		userID         string
		password       string
		mockSetup      func()
		expectedResult bool
	}{
		{
			name:     "Correct_Password",
			userID:   "user1",
			password: password,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT password_hash FROM users WHERE id = $1`)).
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).
						AddRow(string(hashedPassword)))
			},
			expectedResult: true,
		},
		{
			name:     "Incorrect_Password",
			userID:   "user1",
			password: "wrongpassword",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT password_hash FROM users WHERE id = $1`)).
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).
						AddRow(string(hashedPassword)))
			},
			expectedResult: false,
		},
		{
			name:     "User_Not_Found",
			userID:   "user1",
			password: password,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT password_hash FROM users WHERE id = $1`)).
					WithArgs("user1").
					WillReturnError(sql.ErrNoRows)
			},
			expectedResult: false,
		},
		{
			name:     "DB_Error",
			userID:   "user1",
			password: password,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT password_hash FROM users WHERE id = $1`)).
					WithArgs("user1").
					WillReturnError(errors.New("db error"))
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result := storage.CheckPassword(tt.userID, tt.password)

			assert.Equal(t, tt.expectedResult, result)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
