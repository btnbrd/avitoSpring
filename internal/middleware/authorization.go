package middleware

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type Authorizer struct {
	jwtService services.JWTServiceInterface
}

func NewAuthorizer(jwt services.JWTServiceInterface) *Authorizer {
	return &Authorizer{jwtService: jwt}
}

func (h *Authorizer) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.Error{"Authorization header is required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, models.Error{"Authorization header must start with 'Bearer '"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		role, err := h.jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.Error{err.Error()})
			c.Abort()
			return
		}

		c.Set("role", role)
		c.Next()
	}
}
