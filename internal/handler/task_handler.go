package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mini-card-game/internal/middleware"
	"mini-card-game/internal/pkg/app"
	apperrors "mini-card-game/internal/pkg/errors"
	"mini-card-game/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

type claimTaskRequest struct {
	TaskID uint64 `json:"task_id" binding:"required"`
}

func (h *TaskHandler) ListDaily(c *gin.Context) {
	playerID := middleware.PlayerID(c)
	tasks, err := h.taskService.ListDaily(playerID)
	if err != nil {
		app.Fail(c, http.StatusInternalServerError, apperrors.CodeInternal, err.Error())
		return
	}
	app.OK(c, tasks)
}

func (h *TaskHandler) Claim(c *gin.Context) {
	var req claimTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "invalid param")
		return
	}

	rewards, err := h.taskService.Claim(middleware.PlayerID(c), req.TaskID)
	if err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, err.Error())
		return
	}
	app.OK(c, rewards)
}
