package storage

import "avitoSpring/internal/models"

type UserStorage interface {
	CreateUser(user *models.User, password string) (string, error)

	GetUserByEmail(email string) (*models.User, error)

	CheckPassword(userID string, password string) bool
}
