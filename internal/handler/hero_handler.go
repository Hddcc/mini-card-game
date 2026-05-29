package handler

import (
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
