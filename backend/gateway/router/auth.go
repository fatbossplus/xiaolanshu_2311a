package router

import (
	"backend/gateway/internal/handler"
	"backend/gateway/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(v1 *gin.RouterGroup) {

	// 无需登录的认证接口
	auth := v1.Group("/auth")
	{
		auth.POST("/register")
		auth.POST("/login", handler.Login)
		auth.POST("/refresh")
	}

	// 需要登录的认证接口
	authRequired := v1.Group("/auth")
	authRequired.Use(middleware.AuthMiddleware())
	{
		authRequired.POST("/logout")
	}
}
