package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mini-card-game/internal/middleware"
	"mini-card-game/internal/pkg/app"
	apperrors "mini-card-game/internal/pkg/errors"
	"mini-card-game/internal/service"
)

type ActivityLotteryHandler struct {
	activityService *service.ActivityLotteryService
}

func NewActivityLotteryHandler(activityService *service.ActivityLotteryService) *ActivityLotteryHandler {
	return &ActivityLotteryHandler{activityService: activityService}
}

func (h *ActivityLotteryHandler) State(c *gin.Context) {
	playerID := middleware.PlayerID(c)
	output, err := h.activityService.State(playerID, c.ClientIP())
	if err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, err.Error())
		return
	}
	app.OK(c, output)
}

func (h *ActivityLotteryHandler) Draw(c *gin.Context) {
	playerID := middleware.PlayerID(c)
	output, err := h.activityService.Draw(c.Request.Context(), playerID, c.ClientIP())
	if err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, err.Error())
		return
	}
	app.OK(c, output)
}

func (h *ActivityLotteryHandler) Records(c *gin.Context) {
	playerID := middleware.PlayerID(c)
	output, err := h.activityService.Records(playerID)
	if err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, err.Error())
		return
	}
	app.OK(c, output)
}
