package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/thienel/go-backend-template/pkg/config"
	apperror "github.com/thienel/go-backend-template/pkg/error"
	"github.com/thienel/go-backend-template/pkg/response"
)

// RateLimiter middleware using Redis
func RateLimiter(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetConfig()
		if cfg == nil || !cfg.RateLimit.Enabled {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)
		ctx := context.Background()

		// Get current count
		count, err := redisClient.Get(ctx, key).Int64()
		if err != nil && err != redis.Nil {
			// Redis error - allow request
			c.Next()
			return
		}

		if count >= int64(cfg.RateLimit.RequestsPerMinute) {
			response.WriteErrorResponse(c, apperror.ErrTooManyRequests)
			c.Abort()
			return
		}

		// Increment counter
		pipe := redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)
		_, _ = pipe.Exec(ctx)

		c.Next()
	}
}
