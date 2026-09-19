package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/avanthika/efootball-backend/internal/models"
	"github.com/avanthika/efootball-backend/internal/repository"
	"github.com/avanthika/efootball-backend/internal/service"
)

type UserHandler struct {
	service     service.UserService
	tournaments service.SpinWheelService
}

func NewUserHandler(service service.UserService, tournaments service.SpinWheelService) *UserHandler {
	return &UserHandler{service: service, tournaments: tournaments}
}

func (h *UserHandler) ListSubadmins(c *gin.Context) {
	subadmins, err := h.service.ListSubadmins(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"subadmins": subadmins})
}

func (h *UserHandler) GetSubadmin(c *gin.Context) {
	id := c.Param("id")

	subadmin, err := h.service.GetSubadmin(c.Request.Context(), id)
	if errors.Is(err, repository.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "subadmin not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	tours, err := h.tournaments.ListMyTournaments(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	spinWheels := make([]models.Tournament, 0)
	auctionTours := make([]models.Tournament, 0)
	for _, t := range tours {
		if t.Type == models.TournamentTypeAuction {
			auctionTours = append(auctionTours, t)
		} else {
			spinWheels = append(spinWheels, t)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"subadmin":     subadmin,
		"spinWheels":   spinWheels,
		"auctionTours": auctionTours,
	})
}

func (h *UserHandler) DeleteSubadmin(c *gin.Context) {
	err := h.service.DeleteSubadmin(c.Request.Context(), c.Param("id"))
	if errors.Is(err, repository.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "subadmin not found"})
		return
	}
	if errors.Is(err, repository.ErrUserHasDependents) {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete: this subadmin still has tournaments. Delete their tournaments first."})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	c.Status(http.StatusNoContent)
}
