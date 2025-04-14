package services

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/storage"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	store     storage.UserStorage
	secretKey string
}

func NewAuthService(store storage.UserStorage) *AuthService {
	return &AuthService{
		secretKey: "my-super-secret-key-12345",
		store:     store,
	}
}

func (s *AuthService) generateJWTToken(role models.Role) (string, error) {
	claims := jwt.MapClaims{
		"role": role,
		"exp":  time.Now().Add(10 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

func (s *AuthService) DummyLogin(role models.Role) (string, error) {
	if role != models.RoleEmployee && role != models.RoleModerator {
		return "", fmt.Errorf("invalid role: %s", role)
	}
	return s.generateJWTToken(role)
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
	token, err := s.generateJWTToken(role)
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
	token, err := s.generateJWTToken(user.Role)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}

func (s *AuthService) ValidateToken(tokenString string) (models.Role, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}
	role, ok := claims["role"].(string)
	if !ok {
		return "", fmt.Errorf("role not found in token")
	}
	if models.Role(role) != models.RoleEmployee && models.Role(role) != models.RoleModerator {
		return "", fmt.Errorf("invalid role in token: %s", role)
	}
	return models.Role(role), nil
}
