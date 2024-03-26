package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	common "jcpd.cn/common/models"
	commonJWT "jcpd.cn/common/utils/jwt"
	"jcpd.cn/talk/api/auth"
	"jcpd.cn/talk/internal/constants"
	"jcpd.cn/talk/pkg/definition"
	grpcService "jcpd.cn/user/pkg/service"
	"net/http"
	"strconv"
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

const Friend = "friend"
const Group = "group"

func UserRelationDecide(ctx *gin.Context, resp *common.Resp, userId uint32, targetId uint32, fORg string) (bool, *common.NormalErr) {
	request := &grpcService.UserRelationDecideRequest{UserId: userId, TargetId: targetId, FORg: fORg}
	isRelated, err := auth.UserServiceClient.IsRelated(context.Background(), request)
	if err != nil {
		normalErr := common.ToNormalErr(err)
		ctx.JSON(http.StatusOK, resp.Fail(normalErr))
		return false, &normalErr
	}
	return isRelated.IsRelated, nil
}

// ConnectServer websocket连接服务器 - 用于消息提示，和网络检测
// api : /talk/ws/connect/server   [get]	LOGIN
func (h *TalkHandler) ConnectServer(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, _ := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}

	a := ctx.Query("a")
	b := ctx.Query("b")

	aa, _ := strconv.Atoi(a)
	bb, _ := strconv.Atoi(b)

	//	test
	decide, normalErr2 := UserRelationDecide(ctx, resp, uint32(aa), uint32(bb), Friend)

	if normalErr2 != nil {
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(decide))

}
