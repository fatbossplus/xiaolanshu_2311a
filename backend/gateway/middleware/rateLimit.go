package middleware

import (
	"backend/pkg/logger"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// 限流参数
// rps：令牌桶每秒生成的令牌数，即每秒最多允许 100 个请求通过
// burst：令牌桶容量，允许瞬间最多积累 200 个令牌，用于应对突发流量
// 原理：每个请求消耗一个令牌，令牌不足时请求被拒绝；
//
//	令牌按 rps 速率持续补充，桶满后不再积累
var (
	limiters sync.Map          // key=IP，value=*rate.Limiter，每个 IP 独立限流
	rps      = rate.Limit(100) // 每秒生成 100 个令牌
	burst    = 200             // 桶最大容量 200，允许突发 200 个并发请求
)

// getLimiter 按 IP 获取对应的限流器
// sync.Map.LoadOrStore 保证并发安全：首次访问时创建，后续复用同一个限流器
func getLimiter(ip string) *rate.Limiter {
	v, _ := limiters.LoadOrStore(ip, rate.NewLimiter(rps, burst))
	return v.(*rate.Limiter)
}

// RateLimitMiddleware 限流中间件（洋葱模型），按 IP 限流
// 使用令牌桶算法：平均速率 100 req/s，允许突发 200 req
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromCtx(c)

		// 【入站】从当前 IP 的令牌桶中取一个令牌
		// Allow() 立即返回，不等待：有令牌则通过，无令牌则拒绝
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
