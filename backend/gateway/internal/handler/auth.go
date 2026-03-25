package handler

import (
	"backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	response.Success(c, nil)
}
