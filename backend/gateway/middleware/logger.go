package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

// LoggerMiddleware 日志中间件（洋葱模型）
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		logger, _ := zap.NewProduction()
		defer logger.Sync()
		logger.Info("中间件日志记录中",
			// Structured context as strongly typed Field values.
			zap.String("url", c.Request.URL.String()),
			zap.Int("attempt", 3),
			zap.Duration("backoff", time.Second),
		)

		//start := time.Now()
		//
		//// 入站：请求进来时记录
		//logger.Info("-->",
		//	zap.String("1234", "3456"),
		//	zap.String("1234", "3456"),
		//	zap.String("1234", "3456"),
		//	zap.String("1234", "3456"),
		//	zap.String("1234", "3456"),
		//)
		c.Next() // 执行后续中间件和 handler

		// 出站：请求处理完后记录
		//logger.Info("<--",
		//	zap.String("method", c.Request.Method),
		//	zap.String("path", c.Request.URL.Path),
		//	zap.Int("status", c.Writer.Status()),
		//	zap.Duration("latency", time.Since(start)),
		//)
	}
}
