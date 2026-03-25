package middleware

import (
	"backend/pkg/logger"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PermissionMiddleware 权限中间件（洋葱模型）
// 在 AuthMiddleware 之后执行，userID 已由 AuthMiddleware 写入 context
func PermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromCtx(c)

		// 【入站】检查当前用户是否有访问该路由的权限
		userID := c.GetInt64("userID")
		path := c.FullPath()

		fmt.Println("userID:", userID, "path:", path)

		if !hasPermission(userID, path) {
			log.Warn("permission denied",
				zap.Int64("userID", userID),
				zap.String("path", path),
			)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"msg": "无权限"})
			return
		}

		log.Info("permission passed", zap.Int64("userID", userID))

		c.Next()

		// 【出站】权限中间件无需出站逻辑
	}
}

// hasPermission 权限校验，后续对接 casbin 或数据库权限表
func hasPermission(userID int64, path string) bool {
	if userID == 0 {
		return true
	}
	// TODO: 接入 userID 与 path 去 user_acess_permissions 表中查询是否有权限
	return true
}
