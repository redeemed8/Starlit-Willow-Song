package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	common "jcpd.cn/common/models"
	"jcpd.cn/user/internal/constants"
	"jcpd.cn/user/internal/models"
	"jcpd.cn/user/internal/models/vo"
	"jcpd.cn/user/pkg/definition"
	"net/http"
)

// ApplyHandler apply路由的处理器 -- 用于管理各种接口的实现
type ApplyHandler struct {
	cache definition.Cache
}

func NewApplyHandler(type_ definition.CacheType) *ApplyHandler {
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
	return &ApplyHandler{cache: cache_}
}

// ApplyToBeFriend 申请添加为好友
// api : /users/apply/tobe/friend [post]
// post_args : {"friend_name":"xxx","introduction":"xxx"} json LOGIN
func (h *ApplyHandler) ApplyToBeFriend(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var applyVo = vo.ApplyVoHelper.NewApplyVo().ApplyFriendVo
	if err := ctx.ShouldBind(&applyVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 参数校验
	if ok := models.JoinApplyUtil.CheckIntroduce(&(applyVo.Introduction)); !ok {
		ctx.JSON(http.StatusOK, definition.IntroduceNotFormat)
		return
	}
	queryFriend, err1 := models.UserInfoDao.GetUserByUsername(applyVo.FriendName)
	if err1 != nil && !errors.Is(err1, gorm.ErrRecordNotFound) {
		constants.MysqlErr("根据用户名获取用户信息失败", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryFriend.Username == "" || queryFriend.Id == userClaim.Id {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserNotFound))
		return
	}
	//	3.5 检查是否已经为好友
	exists := models.UserInfoUtil.IdIsExists(queryFriend.FriendList, userClaim.Id)
	if exists {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotAddFriendAgain))
		return
	}
	//	4. 查询要添加的人是否已经存在于申请表中
	var columnMap = make(map[string]interface{})
	columnMap["receiver_id"] = queryFriend.Id //	要添加的好友id
	columnMap["apply_type"] = models.Friend
	columnMap["sender_id"] = userClaim.Id
	queryApply, err2 := models.JoinApplyDao.GetApplyByMap(columnMap)
	if err2 != nil && !errors.Is(err2, gorm.ErrRecordNotFound) {
		constants.MysqlErr("根据 Map获取 apply信息出错", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	5. 如果申请存在，说明已经发过了
	if queryApply.ApplyType != "" {
		//	不可重复申请
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotSendApplyAgain))
		return
	}
	//	6. 没有申请过，可以申请, 应为其创建一个新的apply
	var err89 error
	newApply := models.JoinApply{
		SenderId:     userClaim.Id,
		ReceiverId:   queryFriend.Id,
		ApplyType:    models.Friend,
		Status:       models.Pending,
		Introduction: applyVo.Introduction,
	}
	err89 = models.JoinApplyDao.CreateApply(newApply)
	if err89 != nil && !errors.Is(err89, gorm.ErrRecordNotFound) {
		constants.MysqlErr("创建apply出错", err89)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success("请求已发起，等待对方审核.."))
}
