package handlers

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthHandler struct {
	authService services.AuthServiceInterface
}

func NewAuthHandler(authService services.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{err.Error()})
		return
	}

	token, err := h.authService.Register(req.Email, req.Password, models.Role(req.Role))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"token": token})
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		return
	}

	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) DummyLoginHandler(c *gin.Context) {
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{"role is required"})
		return
	}

	role := models.Role(req.Role)
	token, err := h.authService.DummyLogin(role)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
