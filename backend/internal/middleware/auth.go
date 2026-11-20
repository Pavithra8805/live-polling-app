package middleware

import (
	"net/http"
	"strings"

	"live-polling-app/backend/internal/httpx"
	"live-polling-app/backend/internal/service"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "userID"

func RequireAuth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			httpx.Error(c, http.StatusUnauthorized, "missing_token", "a bearer token is required")
			return
		}
		userID, err := auth.ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			httpx.Error(c, http.StatusUnauthorized, "invalid_token", "the access token is invalid or expired")
			return
		}
		c.Set(UserIDKey, userID)
		c.Next()
	}
}
