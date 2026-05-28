package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"mini-card-game/internal/pkg/app"
	apperrors "mini-card-game/internal/pkg/errors"
	jwtpkg "mini-card-game/internal/pkg/jwt"
)

const (
	ContextUserID   = "user_id"
	ContextPlayerID = "player_id"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			app.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "missing token")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "invalid token format")
			c.Abort()
			return
		}

		claims, err := jwtpkg.Parse(secret, parts[1])
		if err != nil {
			app.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "invalid token")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextPlayerID, claims.PlayerID)
		c.Next()
	}
}

func PlayerID(c *gin.Context) uint64 {
	value, exists := c.Get(ContextPlayerID)
	if !exists {
		return 0
	}
	playerID, ok := value.(uint64)
	if !ok {
		return 0
	}
	return playerID
}
