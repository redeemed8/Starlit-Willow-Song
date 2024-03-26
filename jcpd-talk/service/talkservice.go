package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	common "jcpd.cn/common/models"
	commonJWT "jcpd.cn/common/utils/jwt"
	"jcpd.cn/talk/api/auth"
	"jcpd.cn/talk/internal/constants"
	"jcpd.cn/talk/pkg/definition"
	grpcService "jcpd.cn/user/pkg/service"
	"log"
	"net/http"
	"time"
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
	normalErr, curUserClaim := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 升级连接为 websocket连接
	conn, err1 := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	defer func(conn *websocket.Conn) {
		if err := conn.Close(); err != nil {
			log.Println(constants.Err("websocket conn failed to close , caused by : " + err.Error()))
		}
		//	 连接关闭后，将其从记录组中删除
		ConnManager.DelClientConn(curUserClaim.Id, ConnManager.ServerKey)
	}(conn)
	if err1 != nil {
		log.Println(constants.Err("websocket failed to create conn , caused by : " + err1.Error()))
		return
	}
	//	3. 将连接添加到记录组
	ConnManager.AddClientConn(curUserClaim.Id, ConnManager.ServerKey, conn)
	//	4. 服务器响应消息ok
	if err2 := conn.WriteMessage(websocket.TextMessage, []byte("ok")); err2 != nil {
		log.Println(constants.Err("websocket with server failed to send a message that means ready , caused by : " + err2.Error()))
		return
	}
	//	5. 开启持续监听
	timer := time.NewTimer(constants.WebsocketTimeout)
	done := make(chan struct{})

	go func() {
		for {
			//	持续监听其消息，做一个消息的提示
			msgType, msg, err := conn.ReadMessage()
			timer.Reset(constants.WebsocketTimeout)
			if err != nil {
				_ = conn.WriteMessage(websocket.TextMessage, []byte("连接异常，已断开"))
				break
			}
			//	是否是关闭信号
			if msgType == websocket.CloseMessage {
				break
			}
			//	将消息推送给当前用户
			if err123 := conn.WriteMessage(websocket.TextMessage, msg); err123 != nil {
				log.Println(constants.Err("websocket server failed to send tip message , cause by : " + err123.Error()))
				break
			}
		}
		<-done
	}()

	select {
	case <-timer.C:
		_ = conn.WriteMessage(websocket.CloseMessage, []byte("连接已断开"))
		return
	case <-done:
		return
	}

}

////	test
//decide, normalErr2 := UserRelationDecide(ctx, resp, uint32(aa), uint32(bb), Friend)
//
//if normalErr2 != nil {
//return
//}
