package service

import (
	"backend/pkg/response"
	"gateway/internal/request"
	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.CodeParamError)
		return
	}

	if req.LoginType == 1 { // 密码登录

		if req.Password == "" {
			response.Error(c, response.CodeParamError)
			return
		}
		// 查询用户信息，手机号

	} else if req.LoginType == 2 { // 验证码登录
		if req.Code == "" {
			response.Error(c, response.CodeParamError)
		}
		// 查询用户信息，手机号

	} else { // 参数错误
		response.Error(c, response.CodeParamError)
	}

	response.Success(c, nil)
}

func Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.CodeParamError)
		return
	}
	// 验证码是否正确，销毁验证码
	// 手机号是否已注册
	// 保存用户信息

}
