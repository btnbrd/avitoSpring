package handlers

import (
	"avitoSpring/internal/models"
	"avitoSpring/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ReceptionHandler struct {
	receptionService *services.ReceptionService
	authHandler      *AuthHandler
}

func NewReceptionHandler(receptionService *services.ReceptionService, authHandler *AuthHandler) *ReceptionHandler {
	return &ReceptionHandler{
		receptionService: receptionService,
		authHandler:      authHandler,
	}
}

func (h *ReceptionHandler) ReceptionHandler(c *gin.Context) {
	role, _ := c.Get("role")
	if role != models.RoleEmployee {
		c.JSON(http.StatusForbidden, models.Error{
			Message: "Only employees can create receptions",
		})
		return
	}

	var req struct {
		PVZID string `json:"pvzId" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		return
	}

	reception := &models.Reception{
		PVZID: req.PVZID,
	}

	receptionID, err := h.receptionService.CreateReception(reception)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{err.Error()})
		return
	}

	reception.ID = receptionID
	c.JSON(http.StatusCreated, reception)
}

// /pvz/{pvzId}/close_last_reception (POST)
func (h *ReceptionHandler) CloseLastReceptionHandler(c *gin.Context) {
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
			Message: "Only employees can close receptions",
		})
		return
	}

	err := h.receptionService.CloseLastReception(pvzID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		return
	}

	reception, err := h.receptionService.GetLastReceptionByPVZID(pvzID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		return
	}
	if reception == nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "no reception found for PVZ",
		})
		return
	}

	c.JSON(http.StatusOK, reception)
}
