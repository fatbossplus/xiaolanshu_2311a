package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome health")
	})

	v1 := r.Group("/v1") // 版本路由

	// 洋葱模型

	RegisterAuthRouter(v1) // 认证路由
	RegisterUserRouter(v1) // 用户路由
	RegisterNoteRouter(v1) // 笔记路由
	return r
}
