package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

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

func validationErrorMessage(err error) string {
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			switch e.Field() {
			case "Username":
				if e.Tag() == "required" {
					return "请输入账号"
				}
				return "账号长度应为 3 到 64 个字符"
			case "Password":
				if e.Tag() == "required" {
					return "请输入密码"
				}
				return "密码长度应为 6 到 64 个字符"
			case "Nickname":
				if e.Tag() == "required" {
					return "请输入昵称"
				}
				return "昵称长度应为 1 到 64 个字符"
			}
		}
	}
	return "参数格式不正确，请检查账号、密码和昵称"
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, validationErrorMessage(err))
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
		app.Fail(c, http.StatusBadRequest, apperrors.CodeInvalidParam, validationErrorMessage(err))
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
