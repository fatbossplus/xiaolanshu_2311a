package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response 统一响应结构体
// 工单要求: 统一响应格式：{code, message, data, trace_id}
type Response struct {
	Code    int         `json:"code"`               // 业务状态码
	Message string      `json:"message"`            // 提示信息
	Data    interface{} `json:"data,omitempty"`     // 响应数据
	TraceID string      `json:"trace_id,omitempty"` // 链路追踪ID
}

// 统一错误码定义
// 工单要求: 实现统一错误码管理
const (
	CodeSuccess      = 0   // 成功
	CodeParamError   = 400 // 参数错误
	CodeUnauthorized = 401 // 未认证
	CodeForbidden    = 403 // 无权限
	CodeNotFound     = 404 // 资源不存在
	CodeRateLimited  = 429 // 请求过于频繁
	CodeServerError  = 500 // 服务器错误
	CodeServiceBusy  = 503 // 服务繁忙
)

// 错误码对应的消息
var codeMessages = map[int]string{
	CodeSuccess:      "success",
	CodeParamError:   "参数错误",
	CodeUnauthorized: "未登录或Token已过期",
	CodeForbidden:    "无权限访问",
	CodeNotFound:     "资源不存在",
	CodeRateLimited:  "请求过于频繁，请稍后再试",
	CodeServerError:  "服务器内部错误",
	CodeServiceBusy:  "服务繁忙，请稍后再试",
}

// GetCodeMessage 获取错误码对应的消息
func GetCodeMessage(code int) string {

	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应(自定义消息)
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int) {
	c.JSON(400, Response{
		Code:    code,
		Message: GetCodeMessage(code),
	})
}
