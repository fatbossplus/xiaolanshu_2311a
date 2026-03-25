package router

import (
	"backend/gateway/middleware"
	"github.com/gin-gonic/gin"
)

func SetRouter(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "health 。。。。。。")
	})
	v1 := r.Group("/v1")
	// 跨域
	v1.Use(middleware.CORSMiddleware())       // 跨域中间件
	v1.Use(middleware.LoggerMiddleware())     // 日志中间件
	v1.Use(middleware.RateLimitMiddleware())  // 限流中间件
	v1.Use(middleware.PermissionMiddleware()) // 权限中间件
	v1.Use(middleware.TraceMiddleware())      // 跟踪中间件

	// 注册认证路由
	RegisterAuthRoutes(v1)
	RegisterUserRoutes(v1)
}
