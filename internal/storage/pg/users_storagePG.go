package pg

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/storage"
	"database/sql"
	"fmt"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserStoragePG struct {
	DB *sql.DB
}

var _ storage.UserStorage = (*UserStoragePG)(nil)

func NewUserStorage(db *sql.DB) *UserStoragePG {

	return &UserStoragePG{DB: db}
}

func (s *UserStoragePG) CreateUser(user *models.User, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	userID := uuid.New().String()
	query := `
        INSERT INTO users (id, email, password_hash, role)
        VALUES ($1, $2, $3, $4)
        RETURNING id`
	err = s.DB.QueryRow(query, userID, user.Email, string(hash), user.Role).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

func (s *UserStoragePG) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, role FROM users WHERE email = $1`
	err := s.DB.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Role)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (s *UserStoragePG) CheckPassword(userID string, password string) bool {
	var passwordHash string
	query := `SELECT password_hash FROM users WHERE id = $1`
	err := s.DB.QueryRow(query, userID).Scan(&passwordHash)
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	return err == nil
}
