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

type BattleHandler struct {
	battleService *service.BattleService
}

func NewBattleHandler(battleService *service.BattleService) *BattleHandler {
	return &BattleHandler{battleService: battleService}
}

type startBattleRequest struct {
	StageID uint64 `json:"stage_id" binding:"required"`
}

type battleActionRequest struct {
	SessionID uint64 `json:"session_id" binding:"required"`
	Action    string `json:"action" binding:"required"`
	ActorID   string `json:"actor_id"`
	TargetID  string `json:"target_id"`
	SkillID   uint64 `json:"skill_id"`
}

type surrenderBattleRequest struct {
	SessionID uint64 `json:"session_id" binding:"required"`
}

func (h *BattleHandler) Start(c *gin.Context) {
	var req startBattleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "invalid param")
		return
	}

	result, err := h.battleService.Start(middleware.PlayerID(c), req.StageID)
	if err != nil {
		writeBattleError(c, err)
		return
	}
	app.OK(c, result)
}

func (h *BattleHandler) Action(c *gin.Context) {
	var req battleActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "invalid param")
		return
	}

	result, err := h.battleService.Action(middleware.PlayerID(c), service.BattleActionInput{
		SessionID: req.SessionID,
		Action:    req.Action,
		ActorID:   req.ActorID,
		TargetID:  req.TargetID,
		SkillID:   req.SkillID,
	})
	if err != nil {
		writeBattleError(c, err)
		return
	}
	app.OK(c, result)
}

func (h *BattleHandler) Surrender(c *gin.Context) {
	var req surrenderBattleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "invalid param")
		return
	}

	result, err := h.battleService.Surrender(middleware.PlayerID(c), req.SessionID)
	if err != nil {
		writeBattleError(c, err)
		return
	}
	app.OK(c, result)
}

func writeBattleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrStageNotFound):
		app.Fail(c, http.StatusNotFound, apperrors.CodeNotFound, "stage not found")
	case errors.Is(err, service.ErrPrevStageNotCleared):
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "previous stage is not cleared")
	case errors.Is(err, service.ErrNoTeam):
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "please save your team first")
	case errors.Is(err, service.ErrNotEnoughStamina):
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "not enough stamina")
	case errors.Is(err, service.ErrActiveBattleExists):
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "active battle already exists")
	case errors.Is(err, service.ErrBattleConfigMissing):
		app.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, "battle config missing")
	case errors.Is(err, service.ErrBattleNotFound):
		app.Fail(c, http.StatusNotFound, apperrors.CodeNotFound, "battle session not found")
	case errors.Is(err, service.ErrBattleFinished):
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "battle session finished")
	case errors.Is(err, service.ErrBattleExpired):
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "battle session expired")
	case errors.Is(err, service.ErrInvalidBattleAction):
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "invalid battle action")
	default:
		app.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, err.Error())
	}
}
