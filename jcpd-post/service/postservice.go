package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"jcpd.cn/post/internal/constants"
	"jcpd.cn/post/pkg/definition"
)

// PostHandler post路由的处理器 -- 用于管理各种接口的实现
type PostHandler struct {
	cache definition.Cache
	errs  constants.MysqlErr_
}

func NewPostHandler(type_ definition.CacheType) *PostHandler {
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
	return &PostHandler{cache: cache_}
}

func (*PostHandler) A(ctx *gin.Context) {
	ctx.JSON(200, "tttt.....")
}
