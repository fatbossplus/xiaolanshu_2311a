package router

import (
	"backend/gateway/internal/service"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRouter(v1 *gin.RouterGroup) {
	v1.POST("/login", service.Login)       // 登录
	v1.POST("/register", service.Register) // 注册
}
