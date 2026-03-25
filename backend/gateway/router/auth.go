package router

import (
	"backend/gateway/internal/handler"
	"backend/gateway/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(v1 *gin.RouterGroup, h *handler.Handler) {
	auth := v1.Group("/auth")
	{
		auth.POST("/register")
		auth.POST("/login", h.Login)
		auth.POST("/refresh")
	}

	authRequired := v1.Group("/auth")
	authRequired.Use(middleware.AuthMiddleware())
	{
		authRequired.POST("/logout")
	}
}
