package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"live-polling-app/backend/internal/httpx"
)

func RedisRateLimit(client *redis.Client, limit int, window time.Duration, prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity := c.ClientIP()
		if userID, exists := c.Get(UserIDKey); exists {
			if value, ok := userID.(string); ok && value != "" {
				identity = value
			}
		}
		key := fmt.Sprintf("ratelimit:%s:%s", prefix, identity)

		count, err := client.Incr(c.Request.Context(), key).Result()
		if err != nil {
			httpx.Error(c, http.StatusInternalServerError, "rate_limit_error", "could not validate request rate")
			return
		}
		if count == 1 {
			client.Expire(c.Request.Context(), key, window)
		}
		if count > int64(limit) {
			c.Header("Retry-After", fmt.Sprintf("%.0f", window.Seconds()))
			httpx.Error(c, http.StatusTooManyRequests, "rate_limited", "too many requests; please try again later")
			return
		}
		c.Next()
	}
}
