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
	errs  constants.MysqlErr_
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
	normalErr, userClaim := IsLogin(ctx, resp)
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

// GetAllFriendApply 获取所有的好友申请
// api : /users/apply/get-friend/all  [get]  LOGIN
func (h *ApplyHandler) GetAllFriendApply(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 根据 接收人id 获取到所有的 好友apply信息
	columnMap := map[string]interface{}{
		"receiver_id": userClaim.Id,
		"apply_type":  models.Friend,
	}
	applies, err1 := models.JoinApplyDao.GetAppliesByMap(columnMap)
	if h.errs.CheckMysqlErr(err1) {
		constants.MysqlErr("根据 接收人id 获取到所有的 apply信息出错", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if len(applies) == 0 {
		ctx.JSON(http.StatusOK, resp.Success(dto.ApplyInfoDtos{}))
		return
	}
	//	4. 转换成 ApplyInfoDto 的集合
	dtos, err2 := applies.TransToApplyInfoDtos()
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("applies转换为ApplyInfoDtos出错", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	5. 返回
	ctx.JSON(http.StatusOK, resp.Success(dtos))
}

// GetAllGroupApply 获取某个群的 所有申请
// api : /users/apply/get-group/all?groupid=xxx  [get]  LOGIN
func (h *ApplyHandler) GetAllGroupApply(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 获取路径参数
	groupIdStr := ctx.Query("groupid")
	groupIdInt, err1 := strconv.Atoi(groupIdStr)
	if groupIdStr == "" || err1 != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	groupId := uint32(groupIdInt)
	//	3. 根据群 id获取到群信息
	queryGroup, err2 := models.GroupInfoDao.GetGroupInfoById(groupId)
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("根据群id获取群group信息失败", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryGroup.GroupName == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupNotFound))
		return
	}
	//	3.5 校验权限
	if userClaim.Id != queryGroup.LordId && !models.GroupInfoUtil.IsAdmin(queryGroup, userClaim.Id) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserPermissionDenied))
		return
	}
	//	4. 查出所有申请
	columnMap := map[string]interface{}{
		"receiver_id": queryGroup.Id,
		"apply_type":  models.Group,
	}
	applies, err3 := models.JoinApplyDao.GetAppliesByMap(columnMap)
	if h.errs.CheckMysqlErr(err1) {
		constants.MysqlErr("根据 接收人id 获取到所有的 apply信息出错", err3)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if len(applies) == 0 {
		ctx.JSON(http.StatusOK, resp.Success(dto.ApplyInfoDtos{}))
		return
	}
	//	5. 转换成 ApplyInfoDto 的集合
	dtos, err4 := applies.TransToApplyInfoDtos()
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("applies转换为ApplyInfoDtos出错", err4)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	5. 返回
	ctx.JSON(http.StatusOK, resp.Success(dtos))
}

// UpdateApplyStatus 修改申请状态 - 好友邀请
// api : /users/apply/update/friend-status  [post]
// post_args : {"username":"xxx","cur_status":"xxx","to_status":"xxx"}  json  LOGIN
func (h *ApplyHandler) UpdateApplyStatus(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := IsLogin(ctx, resp)
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
	columnMap := map[string]interface{}{
		"sender_id": queryUser.Id, "receiver_id": userClaim.Id, "status": curStatus, "apply_type": models.Friend}
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
		if models.UserInfoUtil.CheckListIsMax(curUser.FriendList) {
			ctx.JSON(http.StatusOK, resp.Fail(definition.FriendEnough))
			return
		}
		//	添加申请人到当前人的好友列表
		models.UserInfoUtil.AddToList(&curUser.FriendList, strconv.Itoa(int(queryUser.Id)))
		columnMap_ := map[string]interface{}{"friend_list": curUser.FriendList}
		_ = models.UserInfoDao.UpdateUserByMap(curUser.Id, columnMap_)
		//	添加当前人到申请人的好友列表
		models.UserInfoUtil.AddToList(&queryUser.FriendList, strconv.Itoa(int(curUser.Id)))
		columnMap_["friend_list"] = queryUser.FriendList
		_ = models.UserInfoDao.UpdateUserByMap(queryUser.Id, columnMap_)
	}
	retMap := map[string]string{"cur_status_ret": toStatus.ToString(), "update_ret": "申请信息已修改"}
	ctx.JSON(http.StatusOK, retMap)
}

// ApplyToGroup 申请加入某个群聊
// api : /users/apply/toadd/group  [post]
// post_args : {"group_id":xxx,"introduction":"xxx"}  json  LOGIN
func (h *ApplyHandler) ApplyToGroup(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var applyVo = vo.ApplyVoHelper.NewApplyVo().ApplyGroupVo
	if err := ctx.ShouldBind(&applyVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 参数校验
	if ok := models.JoinApplyUtil.CheckGroupIntroduce(&applyVo.Introduction); !ok {
		ctx.JSON(http.StatusOK, resp.Fail(definition.IntroduceNotFormat))
		return
	}
	queryGroup, err1 := models.GroupInfoDao.GetGroupInfoById(applyVo.GroupId)
	if h.errs.CheckMysqlErr(err1) {
		constants.MysqlErr("根据id获取群group信息出错", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryGroup.GroupName == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupNotFound))
		return
	}
	//	4. 检查该人是否已经是该群的成员了 或者 是否存在于黑名单中
	if models.GroupInfoUtil.IsMember(queryGroup, userClaim.Id) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.AlreadyIsMember))
		return
	}
	if models.GroupInfoUtil.IsExistBlackList(queryGroup, userClaim.Id) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.AlreadyExistBlackList))
		return
	}
	//	5. 检查以前是否已经发送过申请，但是还未被处理
	apply := models.JoinApply{
		SenderId:   userClaim.Id,
		ReceiverId: queryGroup.Id,
		ApplyType:  models.Group,
		Status:     models.Pending,
	}
	queryApply, err2 := models.JoinApplyDao.GetApplyByInfo(apply)
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("获取历史群申请记录时出错", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryApply.ApplyType != "" {
		//	不可重复申请
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotSendApplyAgain))
		return
	}
	//	6. 有申请资格 但是没申请过，添加一个申请
	apply.Introduction = applyVo.Introduction
	err3 := models.JoinApplyDao.CreateApply(apply)
	if h.errs.CheckMysqlErr(err3) {
		constants.MysqlErr("添加群申请记录时出错", err3)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success("请求已发起，等待对方审核.."))
}

// UpdateApplyGroupStatus 修改加群申请状态
// api : /users/apply/update/group-status  [post]
// post_api : {"username":"xxx","group_id":xxx,"cur_status":"xxx","to_status":"xxx"}  json  LOGIN
func (h *ApplyHandler) UpdateApplyGroupStatus(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var updateVo = vo.ApplyVoHelper.NewApplyVo().UdtApplyGroupStatusVo
	if err := ctx.ShouldBind(&updateVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 参数校验
	//	校验用户名
	if ok := models.UserInfoUtil.CheckUsername(updateVo.Username); !ok {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UnameNotFormat))
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
	if h.errs.CheckMysqlErr(err1) {
		constants.MysqlErr("根据用户名获取用户信息失败", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryUser.Username == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserNotFound))
		return
	}
	//	5. 获取群信息
	queryGroup, err2 := models.GroupInfoDao.GetGroupInfoById(updateVo.GroupId)
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("根据id获取群group信息失败", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryGroup.GroupName == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupNotFound))
		return
	}
	//	5.5 判断权限
	if userClaim.Id != queryGroup.LordId && !models.GroupInfoUtil.IsAdmin(queryGroup, userClaim.Id) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserPermissionDenied))
		return
	}
	//	6. 修改对应的申请信息
	applyConditions := models.JoinApply{SenderId: queryUser.Id, ReceiverId: queryGroup.Id, Status: curStatus, ApplyType: models.Group}
	applyUpdates := models.JoinApply{Status: toStatus}
	err3 := models.JoinApplyDao.UpdateApplyByInfo(applyConditions, applyUpdates)
	if h.errs.CheckMysqlErr(err3) {
		constants.MysqlErr("修改群申请状态时失败", err3)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if errors.Is(err3, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.ApplyNotFound))
		return
	}
	if toStatus == models.JoinApplyUtil.GetAcceptedStatus() {
		//	7. 在该群的id列表中 添加此人id
		groupInfo := models.GroupInfo{
			MemberIds:    models.GroupInfoUtil.AddToList(&queryGroup.MemberIds, strconv.Itoa(int(queryUser.Id))),
			CurPersonNum: queryGroup.CurPersonNum + 1,
		}
		err4 := models.GroupInfoDao.UpdateGroup(queryGroup.Id, groupInfo)
		if h.errs.CheckMysqlErr(err4) {
			constants.MysqlErr("修改群的成员列表时出错", err4)
			ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
			return
		}
		//	8. 在此人的群列表中添加该群的id
		userInfo := models.UserInfo{GroupList: models.UserInfoUtil.AddToList(&queryUser.GroupList, strconv.Itoa(int(queryGroup.Id)))}
		err5 := models.UserInfoDao.UpdateUser(queryUser.Id, userInfo)
		if h.errs.CheckMysqlErr(err5) {
			constants.MysqlErr("修改用户的所在群列表时出错", err5)
			ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
			return
		}
	}
	retMap := map[string]string{"cur_status_ret": toStatus.ToString(), "update_ret": "申请信息已修改"}
	ctx.JSON(http.StatusOK, retMap)
}
