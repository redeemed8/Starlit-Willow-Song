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
	"strings"
)

// GroupHandler group路由的处理器 -- 用于管理各种接口的实现
type GroupHandler struct {
	cache definition.Cache
	errs  constants.MysqlErr_
}

func NewGroupHandler(type_ definition.CacheType) *GroupHandler {
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
	return &GroupHandler{cache: cache_}
}

// CreateGroup 创建群聊
// api : /users/group/create [post]
// post_args : {"group_name":"xxx","group_post":"xxx","max_person_num":xxx}  json LOGIN
func (h *GroupHandler) CreateGroup(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var createVo = vo.UserVoHelper.NewUserVo().CreateGroupVo
	if err := ctx.ShouldBind(&createVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 参数校验
	if ok := models.GroupInfoUtil.CheckGroupName(&(createVo.GroupName)); !ok {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupNameNotFormat))
		return
	}
	if ok := models.GroupInfoUtil.CheckGroupPost(&(createVo.GroupPost)); !ok {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupPostNotFormat))
		return
	}
	if ok := models.GroupInfoUtil.CheckGroupMaxNum(&(createVo.MaxPersonNum)); !ok {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupMaxNumTidy))
		return
	}
	//	4. 创建群信息
	groupInfo := models.GroupInfo{
		GroupName: createVo.GroupName, GroupPost: createVo.GroupPost,
		LordId: userClaim.Id, AdminIds: ",", MemberIds: "," + strconv.Itoa(int(userClaim.Id)) + ",",
		CurPersonNum: 1, MaxPersonNum: createVo.MaxPersonNum, BlackList: ",",
	}
	err9 := models.GroupInfoDao.CreateGroup(&groupInfo)
	if err9 != nil && !errors.Is(err9, gorm.ErrRecordNotFound) {
		constants.MysqlErr("添加群信息失败", err9)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	5. 获取群id
	map_ := map[string]string{"group_id": strconv.Itoa(int(groupInfo.Id))}
	//  6. 获取群主信息
	curUser, err22 := models.UserInfoDao.GetUserById(userClaim.Id)
	if h.errs.CheckMysqlErr(err22) {
		constants.MysqlErr("根据id查用户出错", err22)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	7. 添加群id到 群主用户的群列表
	models.UserInfoUtil.AddToList(&curUser.GroupList, strconv.Itoa(int(groupInfo.Id)))
	err12 := models.UserInfoDao.UpdateUser(userClaim.Id, models.UserInfo{GroupList: curUser.GroupList})
	if h.errs.CheckMysqlErr(err12) {
		constants.MysqlErr("修改用户群列表失败", err12)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(map_))
}

// GetGroupInfoById 根据 群id 获取群基本信息
// api : /users/group/getinfo/byid?id=xxx  [get]  LOGIN
func (h *GroupHandler) GetGroupInfoById(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, _ := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 获取路径参数
	id, err1 := strconv.Atoi(ctx.Query("id"))
	if err1 != nil || id < 1 {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 根据 id获取 groupinfo
	groupInfo, err2 := models.GroupInfoDao.GetGroupInfoById(uint32(id))
	if err2 != nil && !errors.Is(err2, gorm.ErrRecordNotFound) {
		constants.MysqlErr("根据id获取 groupinfo失败", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if groupInfo.GroupName == "" || groupInfo.Status == models.GroupInfoUtil.GetDeletedStatus() {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupNotFound))
		return
	}
	//	4. 封装成 dto进行返回
	ctx.JSON(http.StatusOK, resp.Success(models.GroupInfoUtil.TransToDto(groupInfo)))
}

// GetGroupByName	根据群名查群
// api : /users/group/search?name=xxx  [get]  LOGIN
func (h *GroupHandler) GetGroupByName(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, _ := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 获取路径参数
	groupName := ctx.Query("name")
	if ok := models.GroupInfoUtil.CheckGroupName(&groupName); !ok {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupNameNotFormat))
		return
	}
	//	3. 查数据库
	groupInfos, err1 := models.GroupInfoDao.GetGroupInfoByName(groupName)
	if h.errs.CheckMysqlErr(err1) {
		constants.MysqlErr("根据群名获取群失败", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if len(*groupInfos) < 1 {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupNotFound))
		return
	}
	//	4.  封装为 dto返回
	ctx.JSON(http.StatusOK, resp.Success(models.GroupInfoUtil.TransToDtos(*groupInfos.RemoveDeleted())))
}

// UpdateGroupInfo 修改群基本信息
// api : /users/group/update/info  [post]
// post_args : {"id":xxx,"group_name":"xxx","group_post":"xxx","max_person_num":xxx}  json  LOGIN
func (h *GroupHandler) UpdateGroupInfo(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var updateVo = vo.UserVoHelper.NewUserVo().UpdateGroupInfoVo
	if err := ctx.ShouldBind(&updateVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 获取群信息
	queryGroup, err1 := models.GroupInfoDao.GetGroupInfoById(uint32(updateVo.Id))
	if h.errs.CheckMysqlErr(err1) {
		constants.MysqlErr("根据id获取群信息失败", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryGroup.GroupName == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupNotFound))
		return
	}
	if userClaim.Id != queryGroup.LordId && !models.GroupInfoUtil.IsAdmin(queryGroup, userClaim.Id) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NoGroupAdmin))
		return
	}
	//	4. 参数校验 和 保存
	var groupInfo models.GroupInfo
	if models.GroupInfoUtil.CheckGroupName(&(updateVo.GroupName)) {
		groupInfo.GroupName = updateVo.GroupName
	}
	if models.GroupInfoUtil.CheckGroupPost(&(updateVo.GroupPost)) {
		groupInfo.GroupPost = updateVo.GroupPost
	}
	if models.GroupInfoUtil.CheckGroupMaxNum(&(updateVo.MaxPersonNum)) {
		groupInfo.MaxPersonNum = updateVo.MaxPersonNum
	}
	//	4. 更新
	err9 := models.GroupInfoDao.UpdateGroup(uint32(updateVo.Id), groupInfo)
	if h.errs.CheckMysqlErr(err9) {
		constants.MysqlErr("更新群基本信息失败", err9)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success("群信息已经更新"))
}

// GetJoinedGroup 获取自己的群聊列表
// api : /users/group/getlist  [get]  LOGIN
func (h *GroupHandler) GetJoinedGroup(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 获取用户信息
	curUser, err1 := models.UserInfoDao.GetUserById(userClaim.Id)
	if h.errs.CheckMysqlErr(err1) {
		constants.MysqlErr("根据id获取用户信息失败", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	groupIds := models.UserInfoUtil.TransToUint32Arr(strings.Split(curUser.GroupList, ","))
	if len(groupIds) <= 0 {
		ctx.JSON(http.StatusOK, resp.Success(make([]dto.GroupInfoDto, 0)))
		return
	}
	groups, err3 := models.GroupInfoDao.GetGroupsByIds(groupIds)
	if h.errs.CheckMysqlErr(err3) {
		constants.MysqlErr("根据 id获取所有群信息失败", err3)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	3. 过滤掉已被删除的群
	ctx.JSON(http.StatusOK, resp.Success(groups.RemoveDeleted().Names()))
}

// ExitGroup 退出群聊
// api : /users/group/exit [post]
// post_args : {"group_id":xxx}  json  LOGIN
func (h *GroupHandler) ExitGroup(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	// 	2. 绑定参数
	var exitVo = vo.UserVoHelper.NewUserVo().ExitGroupVo
	if err := ctx.ShouldBind(&exitVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 检验群id是否存在
	groupInfo, err1 := models.GroupInfoDao.GetGroupInfoById(exitVo.GroupId)
	if h.errs.CheckMysqlErr(err1) {
		constants.MysqlErr("根据id获取群信息失败", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if groupInfo.GroupName == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupNotFound))
		return
	}
	//	4. 如果是群主退出，则直接解散群聊
	if userClaim.Id == groupInfo.LordId {
		//	将群状态标记为 deleted
		_ = models.GroupInfoDao.UpdateGroupByMap(groupInfo.Id, map[string]interface{}{"status": models.GroupDeleted})
		ctx.JSON(http.StatusOK, resp.Success("该群已被你解散"))
		//	然后做一个定时任务，定时去数据库的 群id列表中删除 已解散的群id
		return
	}
	//	5. 先在用户群列表中删除掉该群的id
	curUser, err2 := models.UserInfoDao.GetUserById(userClaim.Id)
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("根据id获取用户信息失败", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	models.UserInfoUtil.DeleteFromList(&curUser.GroupList, groupInfo.Id)
	err3 := models.UserInfoDao.UpdateUserByMap(curUser.Id, map[string]interface{}{"group_id": curUser.GroupList})
	if h.errs.CheckMysqlErr(err3) {
		constants.MysqlErr("更新用户群列表失败", err3)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	6. 在群管理列表和群成员列表中删除当前用户id
	models.GroupInfoUtil.DeleteFromList(&groupInfo.AdminIds, userClaim.Id)
	models.GroupInfoUtil.DeleteFromList(&groupInfo.MemberIds, userClaim.Id)
	err4 := models.GroupInfoDao.UpdateGroupByMap(groupInfo.Id,
		map[string]interface{}{"admin_ids": groupInfo.AdminIds, "member_ids": groupInfo.MemberIds})
	if h.errs.CheckMysqlErr(err4) {
		constants.MysqlErr("更新群用户列表失败", err4)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, "已退出群聊")
}

// ChooseUserToBeAdmin 将某人设置为管理员
// api : /users/group/set/admin  [post]
// post_api : {"user_id":xxx,"group_id":xxx}  json  LOGIN
func (h *GroupHandler) ChooseUserToBeAdmin(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var toBeAdminVo = vo.UserVoHelper.NewUserVo().ToBeAdminVo
	if err := ctx.ShouldBind(&toBeAdminVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 参数校验
	queryUser, err1 := models.UserInfoDao.GetUserById(toBeAdminVo.UserId) //	校验要成为admin的人是否存在
	if h.errs.CheckMysqlErr(err1) {
		constants.MysqlErr("根据id获取用户信息异常", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryUser.Username == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserNotFound))
		return
	}
	queryGroup, err2 := models.GroupInfoDao.GetGroupInfoById(toBeAdminVo.GroupId) //	校验群是否存在
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("根据id获取 群group信息异常", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryGroup.GroupName == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupNotFound))
		return
	}
	//	4. 判断是否有权限
	if queryGroup.LordId != userClaim.Id {
		ctx.JSON(http.StatusOK, resp.Fail(definition.OnlyLordUpdate))
		return
	}
	//	5. 校验其是否 是群主 或者 已经是管理员了
	if queryGroup.LordId == queryUser.Id || models.GroupInfoUtil.IsAdmin(queryGroup, queryUser.Id) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.AlreadyIsAdmin))
		return
	}
	if !models.GroupInfoUtil.IsMember(queryGroup, queryUser.Id) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.IsNotGroupMember))
		return
	}
	//	6. 添加id至管理员名单串
	err3 := models.GroupInfoDao.UpdateGroupByMap(
		queryGroup.Id, map[string]interface{}{"admin_ids": models.GroupInfoUtil.AddToList(&queryGroup.AdminIds, strconv.Itoa(int(queryUser.Id)))})
	if h.errs.CheckMysqlErr(err3) {
		constants.MysqlErr("更新群管理员信息异常", err3)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}

	ctx.JSON(http.StatusOK, resp.Success("已将其修改为管理员"))
}

// CancelUserAdmin 取消成员的管理员身份
// api : /users/group/cancel/admin  [post]
// post_args : {"user_id":xxx,"group_id":xxx}  json  LOGIN
func (h *GroupHandler) CancelUserAdmin(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var toBeAdminVo = vo.UserVoHelper.NewUserVo().ToBeAdminVo
	if err := ctx.ShouldBind(&toBeAdminVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 参数校验
	queryUser, err1 := models.UserInfoDao.GetUserById(toBeAdminVo.UserId) //	校验要取消admin的人是否存在
	if h.errs.CheckMysqlErr(err1) {
		constants.MysqlErr("根据id获取用户信息异常", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryUser.Username == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserNotFound))
		return
	}
	queryGroup, err2 := models.GroupInfoDao.GetGroupInfoById(toBeAdminVo.GroupId) //	校验群是否存在
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("根据id获取 群group信息异常", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryGroup.GroupName == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupNotFound))
		return
	}
	//	4. 判断是否有权限
	if queryGroup.LordId != userClaim.Id {
		ctx.JSON(http.StatusOK, resp.Fail(definition.OnlyLordUpdate))
		return
	}
	//	5. 校验 user_id是否是管理员
	if !models.GroupInfoUtil.IsMember(queryGroup, queryUser.Id) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.IsNotGroupMember))
		return
	}
	if !models.GroupInfoUtil.IsAdmin(queryGroup, queryUser.Id) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserIsNotAdmin))
		return
	}
	//	6. 将 user_id 从管理员名单串中移除
	err3 := models.GroupInfoDao.UpdateGroupByMap(
		queryGroup.Id, map[string]interface{}{"admin_ids": models.GroupInfoUtil.DeleteFromList(&queryGroup.AdminIds, queryUser.Id)})
	if h.errs.CheckMysqlErr(err3) {
		constants.MysqlErr("更新群管理员信息异常", err3)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success("已撤销其管理员资格"))
}

// KickUserFromGroup 将用户踢出群聊
// api : /users/group/kick  [post]
// post_api : {"kick_user_id":xxx,"group_id":xxx}  json  LOGIN
func (h *GroupHandler) KickUserFromGroup(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var kickVo = vo.UserVoHelper.NewUserVo().KickUserFromGroupVo
	if err := ctx.ShouldBind(&kickVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 参数校验
	queryUser, err1 := models.UserInfoDao.GetUserById(kickVo.KickUserId) //	校验要踢出的用户是否存在
	if h.errs.CheckMysqlErr(err1) {
		constants.MysqlErr("根据id获取用户信息异常", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryUser.Username == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserNotFound))
		return
	}
	queryGroup, err2 := models.GroupInfoDao.GetGroupInfoById(kickVo.GroupId) //	校验群是否存在
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("根据id获取 群group信息异常", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryGroup.GroupName == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.GroupNotFound))
		return
	}
	if !models.GroupInfoUtil.IsMember(queryGroup, queryUser.Id) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.IsNotGroupMember))
		return
	}
	//	4. 判断是否有权限
	permission := userClaim.Id == queryGroup.LordId
	if !permission && models.GroupInfoUtil.IsAdmin(queryGroup, userClaim.Id) {
		permission = true
	}
	if !permission {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserPermissionDenied)) //	过滤掉普通成员
		return
	}
	if queryGroup.LordId == queryUser.Id {
		ctx.JSON(http.StatusOK, resp.Fail(definition.LordNotKicked)) //	群主无法被踢出
		return
	}
	ok := userClaim.Id == queryGroup.LordId || (models.GroupInfoUtil.IsAdmin(queryGroup, userClaim.Id) && !models.GroupInfoUtil.IsAdmin(queryGroup, queryUser.Id))
	if !ok {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UserPermissionDenied))
		return
	}
	//	5. 踢出此人并加入黑名单
	groupInfo := models.GroupInfo{
		AdminIds:     models.GroupInfoUtil.DeleteFromList(&queryGroup.AdminIds, queryUser.Id),
		MemberIds:    models.GroupInfoUtil.DeleteFromList(&queryGroup.MemberIds, queryUser.Id),
		CurPersonNum: queryGroup.CurPersonNum - 1,
		BlackList:    models.GroupInfoUtil.AddToList(&queryGroup.BlackList, strconv.Itoa(int(queryUser.Id))),
	}
	if err := models.GroupInfoDao.UpdateGroup(queryGroup.Id, groupInfo); h.errs.CheckMysqlErr(err) {
		constants.MysqlErr("踢出用户时，更新群人员信息异常", err)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	userInfo := models.UserInfo{
		GroupList: models.UserInfoUtil.DeleteFromList(&queryUser.GroupList, queryGroup.Id),
	}
	if err := models.UserInfoDao.UpdateUser(queryUser.Id, userInfo); h.errs.CheckMysqlErr(err) {
		constants.MysqlErr("踢出用户时，更新被踢人员的群列表异常", err)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success("该用户成功被踢出并且加入黑名单"))
}
