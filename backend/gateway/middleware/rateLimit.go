package middleware

import (
	"backend/pkg/logger"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// 每个 IP 对应一个限流器，100 req/s，峰值允许 200
var (
	limiters sync.Map
	rps      = rate.Limit(100)
	burst    = 200
)

func getLimiter(ip string) *rate.Limiter {
	v, _ := limiters.LoadOrStore(ip, rate.NewLimiter(rps, burst))
	return v.(*rate.Limiter)
}

// RateLimitMiddleware 限流中间件（洋葱模型），按 IP 限流
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromCtx(c)

		// 【入站】取当前 IP 的限流器，超限则直接拒绝
		ip := c.ClientIP()
		if !getLimiter(ip).Allow() {
			log.Warn("rate limit exceeded", zap.String("ip", ip))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"msg": "请求过于频繁"})
			return
		}

		c.Next()

		// 【出站】限流中间件无需出站逻辑
	}
}
