package middleware

import (
	"backend/pkg/logger"
	"backend/pkg/response"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// 令牌桶限流原理：
//
//	┌─────────────────────────────────┐
//	│  令牌桶（最多存 burst 个令牌）    │ ← 每秒自动补充 rps 个令牌
//	└────────────────┬────────────────┘
//	                 │ 每个请求消耗 1 个令牌
//	        有令牌 → 请求通过 ✅
//	        无令牌 → 请求拒绝 ❌ 429
//
// rps(rate)：令牌生成速率，每秒补充 100 个，决定平均吞吐上限
// burst：桶容量，最多积累 200 个，决定突发流量上限
//
//	服务刚启动或长时间空闲时，桶会积满 burst 个令牌
//	短时间内可承受最多 burst 个并发请求，之后恢复到 rps 速率
//
// 每个 IP 独立一个令牌桶，互不影响
var (
	limiters sync.Map          // key=IP，value=*rate.Limiter，每个 IP 独立限流
	rps      = rate.Limit(100) // 每秒生成 100 个令牌（平均速率上限）
	burst    = 200             // 桶容量 200（突发流量上限）
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
		log := logger.FromCtx(c) // 从上下文获取日志记录器，上文 logger 初始化时已经设置到上下文了，这里可以直接获取

		// 【入站】从当前 IP 的令牌桶中取一个令牌
		// Allow() 立即返回，不等待：有令牌则通过，无令牌则拒绝
		ip := c.ClientIP() // 获取客户端 IP
		if !getLimiter(ip).Allow() {
			log.Warn("rate limit exceeded", zap.String("ip", ip))
			// c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"msg": "请求过于频繁"})
			response.Error(c, response.CodeRateLimited)
			return
		}

		c.Next()

		// 【出站】限流中间件无需出站逻辑
	}
}
