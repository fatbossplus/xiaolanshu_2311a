package middleware

import (
	"backend/pkg/logger"
	"backend/pkg/response"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var jwtSecret = []byte("xiaolanshu_secret") // 实际项目从配置读取

type Claims struct {
	UserID   int64  `json:"userID"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthMiddleware JWT 鉴权中间件（洋葱模型）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromCtx(c)

		// 【入站】从请求头取 token，格式：Authorization: Bearer <token>
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" {
			log.Warn("auth failed: missing token")
			response.Error(c, response.CodeUnauthorized)
			return
		}

		claims, err := parseToken(token)
		if err != nil {
			log.Warn("auth failed: invalid token", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "token 无效"})
			return
		}

		// 将用户信息写入 context，供后续 handler 取用
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		log.Info("auth passed", zap.Int64("userID", claims.UserID))

		c.Next()

		// 【出站】鉴权无需出站逻辑
	}
}

func parseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return token.Claims.(*Claims), nil
}
