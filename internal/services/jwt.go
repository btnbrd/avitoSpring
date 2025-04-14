package services

import (
	"avitoSpring/internal/models"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

type JWTService struct {
	secretKey string
}

type JWTServiceInterface interface {
	GenerateToken(role models.Role) (string, error)
	ValidateToken(tokenString string) (models.Role, error)
}

func NewJWTService() *JWTService {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	return &JWTService{secretKey: secretKey}
}

func (s *JWTService) GenerateToken(role models.Role) (string, error) {
	if s.secretKey == "" {
		return "", fmt.Errorf("JWT secret key is not set")
	}
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

func (s *JWTService) ValidateToken(tokenString string) (models.Role, error) {
	if s.secretKey == "" {
		return "", fmt.Errorf("JWT secret key is not set")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	//token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
	//	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
	//		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	//	}
	//	return []byte(s.secretKey), nil
	//})
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
