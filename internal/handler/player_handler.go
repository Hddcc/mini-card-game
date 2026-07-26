package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

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

type updatePlayerNameRequest struct {
	Name string `json:"name"`
}

func (h *PlayerHandler) UpdateName(c *gin.Context) {
	var req updatePlayerNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "invalid param")
		return
	}

	profile, err := h.playerService.UpdateName(middleware.PlayerID(c), req.Name)
	if err != nil {
		switch {
		case service.IsPlayerNameUpdateClientError(err):
			app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, err.Error())
		case errors.Is(err, gorm.ErrRecordNotFound):
			app.Fail(c, http.StatusNotFound, apperrors.CodeNotFound, "player not found")
		default:
			app.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, "修改失败，请稍后重试")
		}
		return
	}
	app.OK(c, profile)
}
