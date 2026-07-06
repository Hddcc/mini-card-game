package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mini-card-game/internal/middleware"
	"mini-card-game/internal/pkg/app"
	apperrors "mini-card-game/internal/pkg/errors"
	"mini-card-game/internal/service"
)

type HeroHandler struct {
	heroService *service.HeroService
}

func NewHeroHandler(heroService *service.HeroService) *HeroHandler {
	return &HeroHandler{heroService: heroService}
}

func (h *HeroHandler) List(c *gin.Context) {
	playerID := middleware.PlayerID(c)
	heroes, err := h.heroService.List(playerID)
	if err != nil {
		app.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, err.Error())
		return
	}
	app.OK(c, gin.H{"heroes": heroes})
}

type starUpRequest struct {
	PlayerHeroID uint64 `json:"player_hero_id" binding:"required"`
}

func (h *HeroHandler) StarUp(c *gin.Context) {
	var req starUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "invalid param")
		return
	}

	output, err := h.heroService.StarUp(middleware.PlayerID(c), req.PlayerHeroID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrHeroNotOwned):
			app.Fail(c, http.StatusNotFound, apperrors.CodeNotFound, err.Error())
		case errors.Is(err, service.ErrHeroMaxStar),
			errors.Is(err, service.ErrHeroShardNotEnough),
			errors.Is(err, service.ErrHeroGoldNotEnough),
			errors.Is(err, service.ErrHeroStarCostMissing):
			app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, err.Error())
		default:
			app.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, err.Error())
		}
		return
	}
	app.OK(c, output)
}
