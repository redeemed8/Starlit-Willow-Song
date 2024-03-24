package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	common "jcpd.cn/common/models"
	commonJWT "jcpd.cn/common/utils/jwt"
	"jcpd.cn/talk/internal/constants"
	"jcpd.cn/talk/pkg/definition"
	"net/http"
)

// TalkHandler talk路由的处理器 -- 用于管理各种接口的实现
type TalkHandler struct {
	cache definition.Cache
	errs  constants.Err_
}

func NewTalkHandler(type_ definition.CacheType) *TalkHandler {
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
	return &TalkHandler{cache: cache_}
}

// IsLogin 是否登录
func IsLogin(ctx *gin.Context, resp *common.Resp) (*common.NormalErr, commonJWT.UserClaims) {
	userClaims, err := commonJWT.ParseToken(ctx)
	if errors.Is(err, commonJWT.DBException) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerError))
		return &definition.ServerError, userClaims
	}
	if errors.Is(err, commonJWT.NotLoginError) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotLogin))
		return &definition.NotLogin, userClaims
	}
	if err != nil {
		normalErr := common.ToNormalErr(err)
		ctx.JSON(http.StatusOK, resp.Fail(normalErr))
		return &normalErr, userClaims
	}
	return nil, userClaims
}

func (h *TalkHandler) Hello(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "hello world...")
}
