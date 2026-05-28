package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mini-card-game/internal/middleware"
	"mini-card-game/internal/pkg/app"
	apperrors "mini-card-game/internal/pkg/errors"
	"mini-card-game/internal/service"
)

type PlayerHandler struct {
	playerService *service.PlayerService
}

func NewPlayerHandler(playerService *service.PlayerService) *PlayerHandler {
	return &PlayerHandler{playerService: playerService}
}

func (h *PlayerHandler) Profile(c *gin.Context) {
	playerID := middleware.PlayerID(c)
	profile, err := h.playerService.Profile(playerID)
	if err != nil {
		app.Fail(c, http.StatusNotFound, apperrors.CodeNotFound, "player not found")
		return
	}
	app.OK(c, profile)
}

func (h *PlayerHandler) Assets(c *gin.Context) {
	playerID := middleware.PlayerID(c)
	assets, err := h.playerService.Assets(playerID)
	if err != nil {
		app.Fail(c, http.StatusNotFound, apperrors.CodeNotFound, "asset not found")
		return
	}
	app.OK(c, assets)
}
