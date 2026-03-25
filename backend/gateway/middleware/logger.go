package middleware

import (
	"backend/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerMiddleware 日志中间件（洋葱模型）
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 入站：请求进来时记录
		logger.Info("-->",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)

		logger.Debug("trace start",
			zap.String("qqqqqqqqqqqqqq", "2222"),
		)

		c.Next() // 执行后续中间件和 handler

		// 出站：请求处理完后记录
		logger.Info("<--",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}
