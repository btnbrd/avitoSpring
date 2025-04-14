package handlers

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ProductHandler struct {
	productService services.ProductServiceInterface
	authHandler    *AuthHandler
}

func NewProductHandler(productService services.ProductServiceInterface, authHandler *AuthHandler) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		authHandler:    authHandler,
	}
}

func (h *ProductHandler) ProductHandler(c *gin.Context) {
	role, _ := c.Get("role")
	if role != models.RoleEmployee {
		c.JSON(http.StatusForbidden, models.Error{"Only employees can create products"})
		return
	}

	var req struct {
		Type  models.ProductType `json:"type" binding:"required"`
		PVZID string             `json:"pvzId" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{err.Error()})
		return
	}

	product := &models.Product{
		Type: req.Type,
	}

	productID, err := h.productService.CreateProduct(product, req.PVZID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{err.Error()})
		return
	}
	product.ID = productID
	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) DeleteLastProductHandler(c *gin.Context) {
	pvzID := c.Param("pvzId")
	if pvzID == "" {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "pvzId is required",
		})
		return
	}

	role, _ := c.Get("role")
	if role != models.RoleEmployee {
		c.JSON(http.StatusForbidden, models.Error{
			Message: "Only employees can delete products",
		})
		return
	}

	err := h.productService.DeleteLastProduct(pvzID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}
