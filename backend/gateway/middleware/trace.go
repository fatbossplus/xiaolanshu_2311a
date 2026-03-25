package middleware

import (
	"backend/pkg/logger"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const TraceIDKey = "X-Trace-Id"

// TraceMiddleware 链路追踪中间件（洋葱模型）
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先取请求头中的 traceID，没有则生成
		traceID := c.GetHeader(TraceIDKey)
		if traceID == "" {
			traceID = genTraceID()
		}

		// 写入响应头，方便客户端追踪
		c.Header(TraceIDKey, traceID)
		// 将携带 traceID 的 logger 注入 context，后续 handler 通过 logger.FromCtx(c) 取用
		// 这样所有业务日志无需手动传 traceID，自动携带
		logger.InjectCtx(c, zap.String("traceID", traceID))

		start := time.Now()
		logger.FromCtx(c).Info("trace start",
			zap.String("path", c.Request.URL.Path),
		)

		c.Next()

		logger.FromCtx(c).Info("trace end",
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

func genTraceID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
