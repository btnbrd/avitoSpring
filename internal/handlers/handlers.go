package handlers

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type Handlers struct {
	authService *services.AuthService
}

func NewHandlers(authService *services.AuthService) *Handlers {
	return &Handlers{authService: authService}
}

func RegisterHandlers(s *services.Server, authService *services.AuthService) {
	h := NewHandlers(authService)
	s.POST("/dummyLogin", h.DummyLoginHandler)
	s.POST("/register", h.RegisterHandler)
	s.POST("/login", h.LoginHandler)
	s.POST("/pvz", h.AuthMiddleware(), h.PvzHandler)
	s.GET("/pvz", h.AuthMiddleware(), h.PvzGetHandler)
	s.POST("/receptions", h.AuthMiddleware(), h.ReceptionHandler)
	s.POST("/products", h.AuthMiddleware(), h.ProductHandler)
}

func (h *Handlers) RegisterHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.Register(req.Email, req.Password, models.Role(req.Role))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handlers) LoginHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handlers) DummyLoginHandler(c *gin.Context) {
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role is required"})
		return
	}

	role := models.Role(req.Role)
	token, err := h.authService.DummyLogin(role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handlers) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must start with 'Bearer '"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		role, err := h.authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("role", role)
		c.Next()
	}
}

func (h *Handlers) PvzHandler(c *gin.Context) {
	role, _ := c.Get("role")
	if role != models.RoleModerator {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only moderators can create PVZ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Этот эндпоинт создает новый ПВЗ. Доступ только для модераторов."})
}

func (h *Handlers) PvzGetHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Этот эндпоинт получает список ПВЗ с возможностью фильтрации по дате и пагинации."})
}

func (h *Handlers) ReceptionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Этот эндпоинт создает новую приемку товаров для ПВЗ."})
}

func (h *Handlers) ProductHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Этот эндпоинт добавляет товар в текущую приемку для ПВЗ."})
}
