package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mini-card-game/internal/middleware"
	"mini-card-game/internal/pkg/app"
	apperrors "mini-card-game/internal/pkg/errors"
	"mini-card-game/internal/service"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

type saveTeamRequest struct {
	Slots []struct {
		Slot         uint8  `json:"slot" binding:"required"`
		PlayerHeroID uint64 `json:"player_hero_id" binding:"required"`
	} `json:"slots" binding:"required"`
}

func (h *TeamHandler) Save(c *gin.Context) {
	var req saveTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "invalid param")
		return
	}

	slots := make([]service.TeamSlot, 0, len(req.Slots))
	for _, slot := range req.Slots {
		slots = append(slots, service.TeamSlot{
			Slot:         slot.Slot,
			PlayerHeroID: slot.PlayerHeroID,
		})
	}

	if err := h.teamService.Save(middleware.PlayerID(c), slots); err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, err.Error())
		return
	}
	app.OK(c, gin.H{"message": "team saved"})
}

func (h *TeamHandler) Get(c *gin.Context) {
	team, err := h.teamService.Get(middleware.PlayerID(c))
	if err != nil {
		app.Fail(c, http.StatusNotFound, apperrors.CodeNotFound, err.Error())
		return
	}
	app.OK(c, team)
}
