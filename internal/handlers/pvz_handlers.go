package handlers

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type PVZHandler struct {
	pvzService  services.PVZServiceInterface
	authHandler *AuthHandler
}

func NewPVZHandler(pvzService services.PVZServiceInterface, authHandler *AuthHandler) *PVZHandler {
	return &PVZHandler{
		pvzService:  pvzService,
		authHandler: authHandler,
	}
}

// POST
func (h *PVZHandler) PvzHandler(c *gin.Context) {
	role, _ := c.Get("role")
	if role != models.RoleModerator {
		c.JSON(http.StatusForbidden, models.Error{
			Message: "Only moderators can create PVZ",
		})
		return
	}

	var req struct {
		City models.City `json:"city" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		return
	}

	pvz := &models.PVZ{
		//RegistrationDate: req.RegistrationDate,
		City: req.City,
	}

	pvzID, err := h.pvzService.CreatePVZ(pvz)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		return
	}

	pvz.ID = pvzID
	c.JSON(http.StatusCreated, pvz)
}

// GET
func (h *PVZHandler) PvzGetHandler(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	page := 1
	pageSize := 10

	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	if ps := c.Query("limit"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 30 {
			pageSize = val
		}
	}

	role, _ := c.Get("role")
	if role != models.RoleEmployee && role != models.RoleModerator {
		c.JSON(http.StatusForbidden, models.Error{
			Message: "Only employees and moderators can access PVZ list",
		})
		return
	}

	pvzDetails, err := h.pvzService.GetPVZsWithDetails(startDate, endDate, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		return
	}

	response := make([]gin.H, 0)
	for _, detail := range pvzDetails {
		if detail == nil || detail.PVZ == nil {
			continue
		}
		receptionsResponse := make([]gin.H, 0)
		for _, r := range detail.Receptions {
			if r == nil || r.Reception == nil {
				continue
			}
			receptionsResponse = append(receptionsResponse, gin.H{
				"reception": r.Reception,
				"products":  r.Products,
			})
		}
		response = append(response, gin.H{
			"pvz":        detail.PVZ,
			"receptions": receptionsResponse,
		})
	}

	c.JSON(http.StatusOK, response)
}
