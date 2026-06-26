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

type StageHandler struct {
	stageService *service.StageService
}

func NewStageHandler(stageService *service.StageService) *StageHandler {
	return &StageHandler{stageService: stageService}
}

type fightStageRequest struct {
	StageID uint64 `json:"stage_id" binding:"required"`
}

func (h *StageHandler) Progress(c *gin.Context) {
	progress, err := h.stageService.Progress(middleware.PlayerID(c))
	if err != nil {
		app.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, err.Error())
		return
	}
	app.OK(c, progress)
}

func (h *StageHandler) Fight(c *gin.Context) {
	var req fightStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "invalid param")
		return
	}

	result, err := h.stageService.Fight(middleware.PlayerID(c), req.StageID)
	if err != nil {
		if errors.Is(err, service.ErrStageNotFound) {
			app.Fail(c, http.StatusNotFound, apperrors.CodeNotFound, "stage not found")
			return
		}
		if errors.Is(err, service.ErrPrevStageNotCleared) {
			app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "previous stage is not cleared")
			return
		}
		if errors.Is(err, service.ErrNoTeam) {
			app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "please save your team first")
			return
		}
		if errors.Is(err, service.ErrNotEnoughStamina) {
			app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "not enough stamina")
			return
		}
		app.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, err.Error())
		return
	}

	app.OK(c, result)
}
