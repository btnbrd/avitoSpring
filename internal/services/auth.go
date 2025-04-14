package services

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/storage"
	"errors"
	"fmt"
	//"time"
	//"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	store      storage.UserStorage
	jwtService JWTServiceInterface
}

type AuthServiceInterface interface {
	DummyLogin(role models.Role) (string, error)
	Register(email, password string, role models.Role) (string, error)
	Login(email, password string) (string, error)
}

func NewAuthService(store storage.UserStorage, jwtService JWTServiceInterface) *AuthService {
	return &AuthService{
		store:      store,
		jwtService: jwtService,
	}
}

func (s *AuthService) DummyLogin(role models.Role) (string, error) {
	if role != models.RoleEmployee && role != models.RoleModerator {
		return "", fmt.Errorf("invalid role: %s", role)
	}
	token, err := s.jwtService.GenerateToken(role)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) Register(email, password string, role models.Role) (string, error) {
	if email == "" || password == "" {
		return "", errors.New("email and password are required")
	}
	if role != models.RoleEmployee && role != models.RoleModerator {
		return "", fmt.Errorf("invalid role: %s", role)
	}
	existingUser, err := s.store.GetUserByEmail(email)
	if err == nil && existingUser != nil {
		return "", errors.New("user with this email already exists")
	}
	token, err := s.jwtService.GenerateToken(role)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	user := &models.User{
		Email: email,
		Role:  role,
	}
	_, err = s.store.CreateUser(user, password)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	return token, nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	if email == "" || password == "" {
		return "", errors.New("email and password are required")
	}
	user, err := s.store.GetUserByEmail(email)
	if err != nil {
		return "", fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return "", errors.New("user not found")
	}
	if !s.store.CheckPassword(user.ID, password) {
		return "", errors.New("invalid password")
	}
	token, err := s.jwtService.GenerateToken(user.Role)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}
