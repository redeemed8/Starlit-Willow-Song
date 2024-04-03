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
	"jcpd.cn/talk/internal/models"
	"jcpd.cn/talk/pkg/definition"
	grpcService "jcpd.cn/user/pkg/service"
	"log"
	"net/http"
	"strconv"
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

// TalkWithFriend  和好友聊天
// api : /talk/ws/connect/friend?auth=xxx&target=xxx  [get]  LOGIN
func (h *TalkHandler) TalkWithFriend(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, curUserClaim := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 获取路径参数
	target, err1 := strconv.Atoi(ctx.Query("target"))
	if err1 != nil || target < 1 {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserNotFound))
		return
	}
	//	3. 检查是否和当前用户是好友关系 - 防止恶意连接请求
	decide, normalErr2 := UserRelationDecide(ctx, resp, curUserClaim.Id, uint32(target), Friend)
	if normalErr2 != nil {
		return
	}
	if !decide {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotFriend))
		return
	}
	//	4. 满足好友关系， 升级连接为websocket连接
	conn, err2 := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	defer func(conn *websocket.Conn) {
		if err := conn.Close(); err != nil {
			log.Println(constants.Err("websocket conn failed to close , caused by : " + err.Error()))
		}
		//	 连接关闭后，将其从记录组中删除
		ConnManager.DelClientConn(curUserClaim.Id, uint32(target))
	}(conn)
	if err2 != nil {
		log.Println(constants.Err("websocket failed to create conn , caused by : " + err2.Error()))
		return
	}
	//	5. 将连接添加到记录组
	ConnManager.AddClientConn(curUserClaim.Id, uint32(target), conn)
	//	6. 检测与服务器之间的通信
	ServerConn := ConnManager.LoadClientConn(curUserClaim.Id, ConnManager.ServerKey)
	if ServerConn == nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("请先连接服务器"))
		return
	}
	//	7. 这里是当用户进入到与另一个用户的对话页面时才使用的接口，所以在这个接口之前会获取到所有的未读消息，那么这里就应该把未读消息数置零
	err7 := models.MessageCounterDao.CountToZero(uint32(target), curUserClaim.Id)
	if h.errs.CheckMysqlErr(err7) {
		log.Println(constants.Err("未读消息数置零出错 , cause by : " + err7.Error()))
	}
	//	8. 因为对方有可能不在线，所以我们要创建一个未读消息记录表，用来存取未读消息的数量
	err8 := models.MessageCounterDao.CreateMessageCounter(curUserClaim.Id, uint32(target))
	if err8 != nil {
		log.Println(constants.Err("创建未读消息数量表出错 , cause by : " + err8.Error()))
		return
	}
	//	9. 开启监听
	timer := time.NewTimer(constants.WebsocketTimeout)
	done := make(chan struct{})

	go func() {
		for {
			//	此连接只用于监听对方的消息，因为互相发消息用的是两个不同的连接
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
			//	检测对方的在线状态
			targetConn := ConnManager.LoadClientConn(uint32(target), curUserClaim.Id)
			if targetConn == nil {
				//	如果对方不在线，我们将消息保存到数据库
				err999 := models.MessageInfoDao.CreateMessage(&models.Message{SenderId: curUserClaim.Id, ReceiverId: uint32(target), Content: string(msg), Status: models.Unread})
				if err999 != nil {
					_ = conn.WriteMessage(websocket.TextMessage, []byte("消息发送失败，服务器异常"))
				}
			} else {
				//	对方在线，将消息推送给对方用户
				if err123 := targetConn.WriteMessage(websocket.TextMessage, msg); err123 != nil {
					log.Println(constants.Err("消息发送异常 , cause by : " + err123.Error()))
					break
				}
				//	然后将消息保存到数据库
				err999 := models.MessageInfoDao.CreateMessage(&models.Message{SenderId: curUserClaim.Id, ReceiverId: uint32(target), Content: string(msg), Status: models.Readed})
				if err999 != nil {
					_ = conn.WriteMessage(websocket.TextMessage, []byte("消息发送失败，服务器异常"))
				}
			}
			//	通知对方有新消息
			targetConnWithServer := ConnManager.LoadClientConn(uint32(target), ConnManager.ServerKey)
			if targetConnWithServer != nil {
				err34 := targetConnWithServer.WriteMessage(websocket.TextMessage, []byte("你有一条新消息,来自"+curUserClaim.Username))
				if err34 != nil {
					fmt.Println(constants.Err("服务器提示信息异常 , cause by = " + err34.Error()))
				}
			}

		}
		_ = conn.WriteMessage(websocket.CloseMessage, []byte("连接已断开"))
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

// GetUnReadMessageCount  获取和某人对话中的未读消息数量
// api : /talk/message/unread/count?auth=xxx&target=xxx  [get]  LOGIN
func (h *TalkHandler) GetUnReadMessageCount(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, curUserClaim := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 获取路径参数
	target, err1 := strconv.Atoi(ctx.Query("target"))
	if err1 != nil || target < 1 {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserNotFound))
		return
	}
	//	3. 因为grpc请求比普通http请求更耗时，所以这里我们不检查他们是不是好友关系，就算不是好友关系，返回0也是没什么问题的
	//	   就是要进行一次count的全表遍历，那也比grpc的请求快得多了，而且可以考虑使用redis缓存来优化不是好友的问题
	counter, err2 := models.MessageCounterDao.GetUnreadCounter(uint32(target), curUserClaim.Id)
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("获取未读消息数量时出错", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	4. 返回未读数量
	ctx.JSON(http.StatusOK, resp.Success(counter))
}

const ReadedMax = 50

// GetHistoryAndUnreadMsg  获取和某人的历史消息和未读消息
// api : /talk/message/info/detail?auth=xxx&target=xxx  [get]  LOGIN
func (h *TalkHandler) GetHistoryAndUnreadMsg(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, curUserClaim := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 获取路径参数
	target, err1 := strconv.Atoi(ctx.Query("target"))
	if err1 != nil || target < 1 {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserNotFound))
		return
	}
	//	3. 检测互相是否是好友
	decide, normalErr2 := UserRelationDecide(ctx, resp, curUserClaim.Id, uint32(target), Friend)
	if normalErr2 != nil {
		return
	}
	if !decide {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotFriend))
		return
	}
	//	4. 是好友，先获取50条历史消息，按时间正向排序
	history, err2 := models.MessageInfoDao.GetMessage(curUserClaim.Id, uint32(target), models.ForwardSort, models.Readed, ReadedMax)
	if errors.Is(err2, models.UnknownMessageStatus) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UnknownMessageStatus))
		return
	}
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("获取历史消息失败", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	5. 获取所有未读消息，按时间正向排序
	unread, err3 := models.MessageInfoDao.GetMessage(curUserClaim.Id, uint32(target), models.ForwardSort, models.Unread, models.ALL)
	if errors.Is(err3, models.UnknownMessageStatus) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UnknownMessageStatus))
		return
	}
	if h.errs.CheckMysqlErr(err3) {
		constants.MysqlErr("获取未读消息失败", err3)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	6. 返回所有消息
	ret := gin.H{"history": history, "unread": unread}
	ctx.JSON(http.StatusOK, resp.Success(ret))
}

// TalkWithGroup  群聊天
// api : /talk/ws/connect/group?auth=xxx&target=xxx  [get]  LOGIN
func (h *TalkHandler) TalkWithGroup(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, curUserClaim := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 获取路径参数
	target, err1 := strconv.Atoi(ctx.Query("target"))
	if err1 != nil || target < 1 {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupNotFound))
		return
	}
	//	3. 检查是否是该群的成员 - 防止恶意连接
	decide, normalErr2 := UserRelationDecide(ctx, resp, curUserClaim.Id, uint32(target), Group)
	if normalErr2 != nil {
		return
	}
	if !decide {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotGroup))
		return
	}
	//	4. 是群成员, 检测其与服务器之间的通信
	ServerConn := ConnManager.LoadClientConn(curUserClaim.Id, ConnManager.ServerKey)
	if ServerConn == nil {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotServerConn))
		return
	}
	//	5. 与服务器正常，将当前连接升级为websocket连接

}
