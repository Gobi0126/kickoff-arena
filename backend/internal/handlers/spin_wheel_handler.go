package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/avanthika/efootball-backend/internal/repository"
	"github.com/avanthika/efootball-backend/internal/service"
)

type SpinWheelHandler struct {
	service service.SpinWheelService
}

func NewSpinWheelHandler(service service.SpinWheelService) *SpinWheelHandler {
	return &SpinWheelHandler{service: service}
}

func requesterID(c *gin.Context) string {
	return c.GetString("userId")
}

func mapSpinWheelError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrTournamentNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "tournament not found"})
	case errors.Is(err, service.ErrInvalidBracketSize):
		c.JSON(http.StatusBadRequest, gin.H{"error": "player count must be an even number between 2 and 64"})
	case errors.Is(err, service.ErrNotYourTournament):
		c.JSON(http.StatusForbidden, gin.H{"error": "you do not have access to this tournament"})
	case errors.Is(err, service.ErrRegistrationClosed):
		c.JSON(http.StatusConflict, gin.H{"error": "registration is closed for this tournament"})
	case errors.Is(err, service.ErrBracketNotFull):
		c.JSON(http.StatusConflict, gin.H{"error": "bracket is not full yet"})
	case errors.Is(err, service.ErrAlreadyStarted):
		c.JSON(http.StatusConflict, gin.H{"error": "this tournament has already started"})
	case errors.Is(err, service.ErrTiedScore):
		c.JSON(http.StatusBadRequest, gin.H{"error": "scores cannot be tied in a knockout match"})
	case errors.Is(err, service.ErrMatchAlreadyComplete):
		c.JSON(http.StatusConflict, gin.H{"error": "this match result has already been submitted"})
	case errors.Is(err, service.ErrMatchNotReady):
		c.JSON(http.StatusConflict, gin.H{"error": "both players for this match are not decided yet"})
	case errors.Is(err, repository.ErrMatchNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
	case errors.Is(err, service.ErrInvalidOrder):
		c.JSON(http.StatusBadRequest, gin.H{"error": "order must contain every registered player exactly once"})
	case errors.Is(err, service.ErrBracketSizeTooSmall):
		c.JSON(http.StatusConflict, gin.H{"error": "player count cannot be less than the number of players already registered"})
	case errors.Is(err, repository.ErrEntryNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
	}
}

type createTournamentRequest struct {
	Name        string `json:"name" binding:"required"`
	BracketSize int    `json:"bracket_size" binding:"required"`
}

func (h *SpinWheelHandler) Create(c *gin.Context) {
	var req createTournamentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and an even player count (2-64) are required"})
		return
	}

	t, err := h.service.CreateTournament(c.Request.Context(), requesterID(c), req.Name, req.BracketSize)
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *SpinWheelHandler) ListMine(c *gin.Context) {
	list, err := h.service.ListMyTournaments(c.Request.Context(), requesterID(c))
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tournaments": list})
}

func (h *SpinWheelHandler) Get(c *gin.Context) {
	t, err := h.service.GetTournament(c.Request.Context(), c.Param("id"), requesterID(c))
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

type updateBracketSizeRequest struct {
	BracketSize int `json:"bracket_size" binding:"required"`
}

func (h *SpinWheelHandler) UpdateBracketSize(c *gin.Context) {
	var req updateBracketSizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bracket_size is required"})
		return
	}

	t, err := h.service.UpdateBracketSize(c.Request.Context(), c.Param("id"), requesterID(c), req.BracketSize)
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *SpinWheelHandler) Delete(c *gin.Context) {
	err := h.service.DeleteTournament(c.Request.Context(), c.Param("id"), requesterID(c))
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SpinWheelHandler) GetPublic(c *gin.Context) {
	t, err := h.service.GetPublicTournament(c.Request.Context(), c.Param("token"))
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

type registerRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone"`
}

func (h *SpinWheelHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	entry, err := h.service.Register(c.Request.Context(), c.Param("token"), req.Name, req.Phone)
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.JSON(http.StatusCreated, entry)
}

func (h *SpinWheelHandler) AddManualEntry(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	entry, err := h.service.AddManualEntry(c.Request.Context(), c.Param("id"), requesterID(c), req.Name, req.Phone)
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.JSON(http.StatusCreated, entry)
}

func (h *SpinWheelHandler) ListEntries(c *gin.Context) {
	entries, err := h.service.ListEntries(c.Request.Context(), c.Param("id"), requesterID(c))
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

func (h *SpinWheelHandler) UpdateEntry(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	entry, err := h.service.UpdateEntry(c.Request.Context(), c.Param("id"), c.Param("entryId"), requesterID(c), req.Name, req.Phone)
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *SpinWheelHandler) DeleteEntry(c *gin.Context) {
	err := h.service.DeleteEntry(c.Request.Context(), c.Param("id"), c.Param("entryId"), requesterID(c))
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SpinWheelHandler) AssignColors(c *gin.Context) {
	entries, err := h.service.AssignColors(c.Request.Context(), c.Param("id"), requesterID(c))
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

type startSpinRequest struct {
	Order []string `json:"order"`
}

func (h *SpinWheelHandler) StartSpin(c *gin.Context) {
	var req startSpinRequest
	// Body is optional — an empty/missing body means "random order".
	_ = c.ShouldBindJSON(&req)

	matches, err := h.service.StartSpin(c.Request.Context(), c.Param("id"), requesterID(c), req.Order)
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"matches": matches})
}

func (h *SpinWheelHandler) GetFixtures(c *gin.Context) {
	matches, err := h.service.GetFixtures(c.Request.Context(), c.Param("id"), requesterID(c))
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"matches": matches})
}

type submitResultRequest struct {
	Player1Score int `json:"player1_score"`
	Player2Score int `json:"player2_score"`
}

func (h *SpinWheelHandler) SubmitResult(c *gin.Context) {
	var req submitResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player1_score and player2_score are required"})
		return
	}

	match, err := h.service.SubmitMatchResult(c.Request.Context(), c.Param("matchId"), requesterID(c), req.Player1Score, req.Player2Score)
	if err != nil {
		mapSpinWheelError(c, err)
		return
	}
	c.JSON(http.StatusOK, match)
}
