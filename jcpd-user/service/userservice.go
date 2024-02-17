package service

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// UserHandler user路由的处理器 -- 用于管理各种接口的实现
type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// GetCaptcha 根据手机号获取验证码
func (h *UserHandler) GetCaptcha(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "success getCaptcha")
}
