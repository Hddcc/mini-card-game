package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mini-card-game/internal/pkg/app"
	apperrors "mini-card-game/internal/pkg/errors"
	"mini-card-game/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Nickname string `json:"nickname" binding:"required,min=1,max=64"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "invalid param")
		return
	}

	userID, playerID, err := h.authService.Register(service.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		app.Fail(c, http.StatusConflict, apperrors.CodeConflict, err.Error())
		return
	}

	app.OK(c, gin.H{
		"user_id":   userID,
		"player_id": playerID,
	})
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, "invalid param")
		return
	}

	output, err := h.authService.Login(service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		app.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, err.Error())
		return
	}

	app.OK(c, gin.H{
		"token":      output.Token,
		"expires_in": output.ExpiresIn,
		"player": gin.H{
			"player_id": output.PlayerID,
			"nickname":  output.Nickname,
			"level":     output.Level,
		},
	})
}
