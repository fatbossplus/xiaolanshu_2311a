package handler

import "backend/gateway/internal/client"

// Handler 持有所有下游服务客户端，供各 handler 方法使用
type Handler struct {
	user *client.UserClient
}

func NewHandler(user *client.UserClient) *Handler {
	return &Handler{user: user}
}
