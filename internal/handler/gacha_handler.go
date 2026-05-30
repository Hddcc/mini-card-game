package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mini-card-game/internal/middleware"
	"mini-card-game/internal/pkg/app"
	apperrors "mini-card-game/internal/pkg/errors"
	"mini-card-game/internal/service"
)

type GachaHandler struct {
	gachaService *service.GachaService
}

func NewGachaHandler(gachaService *service.GachaService) *GachaHandler {
	return &GachaHandler{gachaService: gachaService}
}

type drawRequest struct {
	PoolID uint64 `json:"pool_id" binding:"required"`
	Times  int    `json:"times" binding:"required"`
}

func (h *GachaHandler) State(c *gin.Context) {
	poolID, err := strconv.ParseUint(c.DefaultQuery("pool_id", "1"), 10, 64)
	if err != nil || poolID == 0 {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "invalid param")
		return
	}

	playerID := middleware.PlayerID(c)
	output, err := h.gachaService.State(playerID, poolID)
	if err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, err.Error())
		return
	}

	app.OK(c, output)
}

func (h *GachaHandler) Draw(c *gin.Context) {
	var req drawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "invalid param")
		return
	}

	playerID := middleware.PlayerID(c)
	output, err := h.gachaService.Draw(playerID, req.PoolID, req.Times)
	if err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, err.Error())
		return
	}

	app.OK(c, output)
}
