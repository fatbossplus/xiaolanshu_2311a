package handler

import (
	"backend/pkg/logger"
	"backend/pkg/response"
	userpb "backend/services/user/proto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	log := logger.FromCtx(c)

	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("login param invalid", zap.Error(err))
		response.Error(c, response.CodeParamError)
		return
	}

	// 调用 user 服务 gRPC 登录接口
	resp, err := h.user.Login(c.Request.Context(), &userpb.LoginReq{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		log.Error("user service login failed", zap.Error(err))
		response.Error(c, response.CodeServerError)
		return
	}

	log.Info("login success", zap.Int64("userID", resp.UserId))
	response.Success(c, gin.H{
		"token":  resp.Token,
		"userID": resp.UserId,
	})
}
