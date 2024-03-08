package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"jcpd.cn/post/internal/constants"
	"jcpd.cn/post/pkg/definition"
	"net/http"
)

// SocialHandler social路由的处理器 -- 用于管理各种接口的实现
type SocialHandler struct {
	cache definition.Cache
	errs  constants.Err_
}

func NewSocialHandler(type_ definition.CacheType) *SocialHandler {
	var cache_ definition.Cache
	switch type_ {
	case definition.CacheRedis:
		cache_ = definition.Rc
	case definition.CacheMongo:
		fmt.Println("wait to do...")
	case definition.CacheMysql:
		fmt.Println("wait to do...")
	case definition.Memcahce:
		fmt.Println("wait to do...")
	default:
		cache_ = definition.Rc
	}
	return &SocialHandler{cache: cache_}
}

// Testing 测试...
func (h *SocialHandler) Testing(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "testing...")
}
