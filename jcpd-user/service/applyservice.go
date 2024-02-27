package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	common "jcpd.cn/common/models"
	"jcpd.cn/user/internal/constants"
	"jcpd.cn/user/internal/models"
	"jcpd.cn/user/internal/models/dto"
	"jcpd.cn/user/internal/models/vo"
	"jcpd.cn/user/pkg/definition"
	"net/http"
	"strconv"
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

// GetAllAppliesByStatus 获取所有的申请 - 好友或群邀请
// api : /users/apply/get/all?status=xxx&type=[friend/group]  [get]  LOGIN
func (h *ApplyHandler) GetAllAppliesByStatus(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 获取路径参数
	s := ctx.Query("status")
	status := models.JoinApplyUtil.TransToStatus(s)
	if status == models.JoinApplyUtil.GetDftWebStatus() {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidApplyStatus))
		return
	}
	t := ctx.Query("type")
	if ok := models.JoinApplyUtil.CheckType(t); !ok {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidApplyType))
		return
	}

	//	3. 根据 审核状态status 和 接收人id 获取到所有的 apply信息
	columnMap := map[string]interface{}{
		"status":      status,
		"receiver_id": userClaim.Id,
		"apply_type":  t,
	}
	applies, err1 := models.JoinApplyDao.GetAppliesByMap(columnMap)
	if err1 != nil && !errors.Is(err1, gorm.ErrRecordNotFound) {
		constants.MysqlErr("根据 审核状态status 和 接收人id 获取到所有的 apply信息出错", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if len(applies) == 0 {
		ctx.JSON(http.StatusOK, resp.Success(dto.ApplyInfoDtos{}))
		return
	}
	//	4. 转换成 ApplyInfoDto 的集合
	dtos, err2 := applies.TransToApplyInfoDtos(status)
	if err2 != nil && !errors.Is(err2, gorm.ErrRecordNotFound) {
		constants.MysqlErr("applies转换为ApplyInfoDtos出错", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	5. 返回
	ctx.JSON(http.StatusOK, resp.Success(dtos))
}

// UpdateApplyStatus 修改申请状态 - 好友邀请
// api : /users/apply/update/status  [post]
// post_args : {"username":"xxx","apply_type":"xxx","cur_status":"xxx","to_status":"xxx"}  json  LOGIN
func (h *ApplyHandler) UpdateApplyStatus(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var updateVo = vo.ApplyVoHelper.NewApplyVo().UpdateApplyStatusVo
	if err := ctx.ShouldBind(&updateVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 参数检验
	//	校验用户名
	if ok := models.UserInfoUtil.CheckUsername(updateVo.Username); !ok {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UnameNotFormat))
		return
	}
	//	申请类型
	if ok := models.JoinApplyUtil.CheckType(updateVo.ApplyType); !ok {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidApplyType))
		return
	}
	//	当前审核状态
	curStatus := models.JoinApplyUtil.TransToStatus(updateVo.CurStatus)
	if curStatus != models.JoinApplyUtil.GetPendStatus() {
		ctx.JSON(http.StatusOK, resp.Fail(definition.StatusNotUpdate))
		return
	}
	//	要修改成的审核状态
	toStatus := models.JoinApplyUtil.TransToStatus(updateVo.ToStatus)
	if toStatus != models.AAA && toStatus != models.RRR {
		ctx.JSON(http.StatusOK, resp.Fail(definition.StatusNotToUpdate))
		return
	}
	//	4. 获取到请求人的信息
	queryUser, err1 := models.UserInfoDao.GetUserByUsername(updateVo.Username)
	if err1 != nil && !errors.Is(err1, gorm.ErrRecordNotFound) {
		constants.MysqlErr("根据用户名获取用户信息失败", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryUser.Username == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserNotFound))
		return
	}
	//	5. 获取对应的申请信息
	columnMap := map[string]interface{}{"sender_id": queryUser.Id, "receiver_id": userClaim.Id, "apply_type": updateVo.ApplyType, "status": curStatus}
	apply, err2 := models.JoinApplyDao.GetApplyByMap(columnMap)
	if err2 != nil && !errors.Is(err2, gorm.ErrRecordNotFound) {
		constants.MysqlErr("修改申请状态时获取申请信息失败", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if apply.ApplyType == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.ApplyNotFound))
		return
	}
	//	6. 找到了进行修改即可
	map_ := map[string]interface{}{
		"status": toStatus,
	}
	err3 := models.JoinApplyDao.UpdateApplyByMap(apply.Id, map_)
	if err3 != nil && !errors.Is(err3, gorm.ErrRecordNotFound) {
		constants.MysqlErr("修改审核状态失败", err3)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	7. 添加到其好友列表中
	if toStatus == models.JoinApplyUtil.GetAcceptedStatus() {
		//	获取当前用户信息
		curUser, _ := models.UserInfoDao.GetUserById(userClaim.Id)
		if updateVo.ApplyType == models.Friend {
			//	添加申请人到当前人的好友列表
			models.UserInfoUtil.AddToList(&curUser.FriendList, strconv.Itoa(int(queryUser.Id)))
			columnMap_ := map[string]interface{}{"friend_list": curUser.FriendList}
			_ = models.UserInfoDao.UpdateUserByMap(curUser.Id, columnMap_)
			//	添加当前人到申请人的好友列表
			models.UserInfoUtil.AddToList(&queryUser.FriendList, strconv.Itoa(int(curUser.Id)))
			columnMap_["friend_list"] = queryUser.FriendList
			_ = models.UserInfoDao.UpdateUserByMap(queryUser.Id, columnMap_)
		}
	}
	retMap := map[string]string{"cur_status_ret": toStatus.ToString(), "update_ret": "申请信息已修改"}
	ctx.JSON(http.StatusOK, retMap)
}
