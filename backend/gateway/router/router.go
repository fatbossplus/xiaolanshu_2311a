package router

import (
	"backend/gateway/internal/handler"
	"backend/gateway/middleware"
	"github.com/gin-gonic/gin"
)

func SetRouter(r *gin.Engine, h *handler.Handler) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	v1 := r.Group("/v1")
	v1.Use(middleware.CORSMiddleware())       // 跨域
	v1.Use(middleware.LoggerMiddleware())     // 日志
	v1.Use(middleware.TraceMiddleware())      // 链路追踪
	v1.Use(middleware.RateLimitMiddleware())  // 限流
	v1.Use(middleware.PermissionMiddleware()) // 权限

	RegisterAuthRoutes(v1, h)
	RegisterUserRoutes(v1, h)
}
