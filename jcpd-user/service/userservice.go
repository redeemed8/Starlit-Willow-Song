package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"
	common "jcpd.cn/common/models"
	commonJWT "jcpd.cn/common/utils/jwt"
	"jcpd.cn/user/internal/constants"
	"jcpd.cn/user/internal/models"
	"jcpd.cn/user/internal/models/dto"
	"jcpd.cn/user/internal/models/vo"
	"jcpd.cn/user/internal/options"
	"jcpd.cn/user/pkg/definition"
	"jcpd.cn/user/utils"
	"log"
	"net/http"
	"time"
)

// UserHandler user路由的处理器 -- 用于管理各种接口的实现
type UserHandler struct {
	cache definition.Cache
}

func NewUserHandler(type_ definition.CacheType) *UserHandler {
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
	return &UserHandler{cache: cache_}
}

// GetCaptcha 根据手机号获取验证码,mode为验证码用途,0登录/1修改密码/2绑定手机
// api : /users/getCaptcha?mobile=xxx&mode=xxx	[get]
func (h *UserHandler) GetCaptcha(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 获取路径参数 mobile 和 mode
	mobile := ctx.Query("mobile")
	mode := ctx.Query("mode")
	//	2. 参数校验
	if ok := utils.VerifyMobile(mobile); !ok {
		ctx.JSON(http.StatusOK, resp.Fail(definition.InvalidMobile))
		return
	}
	var keyPrefix string
	if keyPrefix = constants.MatchModeCode(mode); keyPrefix == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.InvalidMode))
		return
	}
	//	3. 查询 redis中是否已经有了验证码 -- 60s过期，未过期不可再发
	key := keyPrefix + mobile
	queryCode, err := h.cache.Get(key)
	if err != nil && err != redis.Nil {
		//	此处说明 redis服务有问题，应当进行降级处理
		constants.RedisErr("获取验证码异常", err)
		//	TODO 降级处理，放入消息队列进行通知各个服务，但这里不做，应在grpc里
		//	...
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryCode != "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.FrequentCaptcha))
		return
	}
	//	4 生成 6位数字的验证码
	code := utils.MakeCodeWithNumber(6, time.Now().Second())
	//	5. 调用短信平台 api，来发送短信
	go func(key_ string) {
		//	模拟耗时
		time.Sleep(2 * time.Second)
		log.Println("Successfully send captcha to mobile : ", mobile)
		//	6. 缓存到 redis
		err1 := h.cache.Put(key_, code, 1*time.Minute)
		if err1 != nil {
			msg := fmt.Sprintf("Failed to save the mobile and captcha to redis : %s : %s , cause by : %v \n", key, code, err1)
			constants.RedisErr(msg, err1)
		}
	}(key)
	ctx.JSON(http.StatusOK, resp.Success("验证码已发送"))
}

// RegisterUser 使用用户名和密码注册新用户
// api : /users/register  [post]
// post_args : {"username":"xxx","password":"xxx","repassword":"xxx"} json
func (h *UserHandler) RegisterUser(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 获取请求体参数
	register := vo.UserVoHelper.NewUserVo().RegisterVo
	if err := ctx.ShouldBind(&register); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	2. 校验两次密码是否一致
	if register.Password != register.Repassword {
		ctx.JSON(http.StatusOK, resp.Fail(definition.PwdNotSame))
		return
	}
	//	3. 密码一致，看用户名 是否合法 和 是否已被使用
	if !models.UserInfoUtil.CheckUsername(register.Username) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UnameNotFormat))
		return
	}
	queryUser, err1 := models.UserInfoDao.GetUserByUsername(register.Username)
	if err1 != nil && !errors.Is(err1, gorm.ErrRecordNotFound) {
		constants.MysqlErr("根据用户名获取用户异常", err1)
		//	TODO 降级处理，放入消息队列进行通知各个服务，但这里不做，应在grpc里
		//	...
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryUser.Username != "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UnameExists))
		return
	}
	//	4. 没被使用，那就可以为其注册
	userinfo := models.UserInfo{
		Username: register.Username,
		Password: utils.Md5Sum(register.Password),
		UUID:     uuid.New().String(),
		Sex:      models.UserInfoUtil.GetDefaultSex(),
	}
	err2 := models.UserInfoDao.CreateUser(userinfo)
	if err2 != nil && !errors.Is(err2, gorm.ErrRecordNotFound) {
		constants.MysqlErr("创建用户失败", err2)
		//	TODO 降级处理，放入消息队列进行通知各个服务，但这里不做，应在grpc里
		//	...
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success("注册成功，请登录"))
}

// LoginMobile 用户 手机号验证码登录 -- 不用注册
// api : /users/login/mobile  [post]
// post_args : {"mobile":"1xxxxxxxxxx","captcha":"xxx"}  json
func (h *UserHandler) LoginMobile(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 绑定参数
	var loginInfo = vo.UserVoHelper.NewUserVo().LoginMobileVo
	if err := ctx.ShouldBind(&loginInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	2. 参数校验，检验 手机号 和 验证码
	if !utils.VerifyMobile(loginInfo.Mobile) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.InvalidMobile))
		return
	}
	if loginInfo.Captcha == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.CaptchaError))
		return
	}
	keyPrefix := constants.MatchModeCode(constants.LoginMode)
	queryCaptcha, err1 := h.cache.Get(keyPrefix + loginInfo.Mobile)
	if err1 != nil && err1 != redis.Nil {
		constants.RedisErr("查询验证码失败", err1)
		//	TODO 降级处理，放入消息队列进行通知各个服务，但这里不做，应在grpc里
		//	...
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryCaptcha == "" || err1 == redis.Nil {
		ctx.JSON(http.StatusOK, resp.Fail(definition.CaptchaNotSend))
		return
	}
	if queryCaptcha != loginInfo.Captcha {
		ctx.JSON(http.StatusOK, resp.Fail(definition.CaptchaError))
		return
	}
	//	3. 验证码正确，判断其是否是第一次登录 - 在数据库查该手机号
	queryUser, err2 := models.UserInfoDao.GetUserByPhone(loginInfo.Mobile)
	if err2 != nil && !errors.Is(err2, gorm.ErrRecordNotFound) {
		constants.MysqlErr("根据手机号获取用户异常", err2)
		//	TODO 降级处理，放入消息队列进行通知各个服务，但这里不做，应在grpc里
		//	...
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	var userClaim commonJWT.UserClaims
	if queryUser.Phone == "" {
		//	第一次登录，为其创建用户信息
		userinfo := models.UserInfo{
			Phone:    loginInfo.Mobile,
			Username: models.UserInfoUtil.GetDefaultName(),
			UUID:     uuid.New().String(),
			Sex:      models.UserInfoUtil.GetDefaultSex(),
		}
		err3 := models.UserInfoDao.CreateUser(userinfo)
		if err3 != nil {
			constants.MysqlErr("创建用户失败", err3)
			//	TODO 降级处理，放入消息队列进行通知各个服务，但这里不做，应在grpc里
			//	...
			ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
			return
		}
		userClaim.Id = userinfo.Id
		userClaim.UUID = userinfo.UUID
	} else {
		//	用户以前登录过，直接赋值
		userClaim.Id = queryUser.Id
		userClaim.UUID = queryUser.UUID
	}
	//	4. 返回 登录token
	token, _ := commonJWT.MakeToken(userClaim)
	ctx.JSON(http.StatusOK, resp.Success(token))
}

// LoginPasswd 用户名密码登录 - 需要注册
// api : /users/login/passwd  [post]
// post_args : {"username":"xxx","password":"xxx"} json
func (h *UserHandler) LoginPasswd(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 绑定参数
	var loginInfo = vo.UserVoHelper.NewUserVo().LoginPasswdVo
	if err := ctx.ShouldBind(&loginInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	2. 参数校验
	if len(loginInfo.Username) > 31 {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UnameNotFound))
		return
	}
	queryUser, err1 := models.UserInfoDao.GetUserByUsername(loginInfo.Username)
	if err1 != nil && !errors.Is(err1, gorm.ErrRecordNotFound) {
		constants.MysqlErr("创建用户失败", err1)
		//	TODO 降级处理，放入消息队列进行通知各个服务，但这里不做，应在grpc里
		//	...
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryUser.Username == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UnameNotFound))
		return
	}
	if queryUser.Password == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.PwdNotSet))
		return
	}
	//	3. 校验密码 md5
	if queryUser.Password != utils.Md5Sum(loginInfo.Password) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.PwdError))
		return
	}
	//	4. 用户名密码用过，发放 token
	token, _ := commonJWT.MakeToken(commonJWT.UserClaims{Id: queryUser.Id, UUID: queryUser.UUID})
	ctx.JSON(http.StatusOK, resp.Success(token))
}

// UserBindMobile 用户绑定手机号，或修改绑定手机号
// api : /users/bind/mobile  [post]
// post_args : {"mobile":"1xxxxxxxxxx","captcha":"xxxxxx"}  json LOGIN
func (h *UserHandler) UserBindMobile(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var bind = vo.UserVoHelper.NewUserVo().BindMobileVo
	if err := ctx.ShouldBind(&bind); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	// 	3. 校验参数 - 手机号和验证码
	if !utils.VerifyMobile(bind.Mobile) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.InvalidMobile))
		return
	}
	if bind.Captcha == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.CaptchaError))
		return
	}
	keyPrefix := constants.MatchModeCode(constants.BindMode)
	queryCaptcha, err1 := h.cache.Get(keyPrefix + bind.Mobile)
	if err1 != nil && err1 != redis.Nil {
		constants.RedisErr("查询验证码失败", err1)
		//	TODO 降级处理，放入消息队列进行通知各个服务，但这里不做，应在grpc里
		//	...
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryCaptcha == "" || err1 == redis.Nil {
		ctx.JSON(http.StatusOK, resp.Fail(definition.CaptchaNotSend))
		return
	}
	if queryCaptcha != bind.Captcha {
		ctx.JSON(http.StatusOK, resp.Fail(definition.CaptchaError))
		return
	}
	//	4. 验证码正确，进行绑定操作
	columnMap := make(map[string]interface{})
	columnMap["phone"] = bind.Mobile
	err9 := models.UserInfoDao.UpdateUserByMap(userClaim.Id, columnMap)
	if err9 != nil && !errors.Is(err9, gorm.ErrRecordNotFound) {
		constants.MysqlErr("绑定用户手机号失败", err9)
		//	TODO 降级处理，放入消息队列进行通知各个服务，但这里不做，应在grpc里
		//	...
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success("手机号已更新"))
}

// GetRepasswdToken	忘记密码/修改密码前置操作 - 申请修改权限
// api : /users/repasswd/check  [post]
// post_args : {"username":"xxx","mobile":"xxx","captcha":"xxx"}  json
func (h *UserHandler) GetRepasswdToken(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 绑定参数
	repwdCheckVo := vo.UserVoHelper.NewUserVo().RepwdCheckVo
	if err := ctx.ShouldBind(&repwdCheckVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	// 2. 初审用户信息
	if ok := utils.VerifyMobile(repwdCheckVo.Mobile); !ok {
		ctx.JSON(http.StatusOK, resp.Fail(definition.InvalidMobile))
		return
	}
	if ok := models.UserInfoUtil.CheckUsername(repwdCheckVo.Username); !ok {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UnameNotFormat))
		return
	}
	//	3. 核实用户手机号和用户名是否属于同一人
	queryUser, err1 := models.UserInfoDao.GetUserByPhone(repwdCheckVo.Mobile)
	if err1 != nil && !errors.Is(err1, gorm.ErrRecordNotFound) {
		constants.MysqlErr("根据手机号查询用户信息失败", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryUser.Username == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UnameNotFound))
		return
	}
	if queryUser.Phone == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotBindMobile))
		return
	}
	if queryUser.Username != repwdCheckVo.Username {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotMatchMobile))
		return
	}
	//	4. 校验验证码
	if repwdCheckVo.Captcha == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.CaptchaError))
		return
	}
	keyPrefix := constants.MatchModeCode(constants.ForgetMode)
	queryCaptcha, err2 := h.cache.Get(keyPrefix + repwdCheckVo.Mobile)
	if err2 != nil && err2 != redis.Nil {
		constants.RedisErr("查询验证码失败", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryCaptcha == "" || err2 == redis.Nil {
		ctx.JSON(http.StatusOK, resp.Fail(definition.CaptchaNotSend))
		return
	}
	if queryCaptcha != repwdCheckVo.Captcha {
		ctx.JSON(http.StatusOK, resp.Fail(definition.CaptchaError))
		return
	}
	//	5. 验证码正确，发送一个 token作为 修改密码的凭据 ,存入redis,因为可以用完就删除
	tidyToken := utils.MakeCodeWithNumber(10, int(queryUser.Id))
	err3 := h.cache.Put(constants.RepwdCheckPrefix+repwdCheckVo.Mobile, tidyToken, constants.RepwdCheckExpire)
	if err3 != nil {
		constants.RedisErr("保存验证码到redis失败:"+repwdCheckVo.Mobile+":"+tidyToken, err3)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(tidyToken))
}

// Repassword 修改密码
// api : /users/repasswd?auth2=xxx  [post]
// post_args : {"mobile":"xxx","password":"xxx","repassword":"xxx"} json TIDY_TOKEN
func (h *UserHandler) Repassword(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 绑定参数
	var repwdVo = vo.UserVoHelper.NewUserVo().RepwdVo
	if err := ctx.ShouldBind(&repwdVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	1.5 简单校验
	if repwdVo.Password != repwdVo.Repassword {
		ctx.JSON(http.StatusOK, resp.Fail(definition.PwdNotSame))
		return
	}
	//	2. 校验路径参数的身份 token
	tidyToken := ctx.Query("auth2")
	if tidyToken == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotAuth2Token))
		return
	}
	//	3. 验证手机号真实性
	queryUser, err2 := models.UserInfoDao.GetUserByPhone(repwdVo.Mobile)
	if err2 != nil && !errors.Is(err2, gorm.ErrRecordNotFound) {
		constants.MysqlErr("根据手机号获取用户信息失败", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryUser.Phone == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.PhoneNotFound))
		return
	}
	//	4. 查 redis 的 tidyToken
	key := constants.RepwdCheckPrefix + repwdVo.Mobile
	queryToken, err1 := h.cache.Get(key)
	if err1 != nil && err1 != redis.Nil {
		constants.RedisErr("从redis获取修改密码的token失败", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if err1 == redis.Nil || queryToken != tidyToken {
		ctx.JSON(http.StatusOK, resp.Fail(definition.Auth2TokenErr))
		return
	}
	//	5. token正确， 修改其密码，以及UUID
	columnMap := make(map[string]interface{})
	columnMap["password"] = utils.Md5Sum(repwdVo.Password)
	columnMap["uuid"] = uuid.New().String()
	err3 := models.UserInfoDao.UpdateUserByMap(queryUser.Id, columnMap)
	if err3 != nil && !errors.Is(err3, gorm.ErrRecordNotFound) {
		constants.MysqlErr("更新用户密码失败", err3)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	6. 立即删除 redis的key
	c, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	options.C.RDB.Del(c, key)
	ctx.JSON(http.StatusOK, resp.Success("密码已更新，请重新登录"))
}

// UpdateUserInfo 修改 用户名、性别、个性签名 - 不想修改的字段传空字符串即可
// api : /users/update/info [post]
// post_args : {"username":"","sex":"","sign":""} json LOGIN
func (h *UserHandler) UpdateUserInfo(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var updateInfo = vo.UserVoHelper.NewUserVo().UpdateUserInfoVo
	if err := ctx.ShouldBind(&updateInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 参数校验
	if updateInfo.Username != "" && !models.UserInfoUtil.CheckUsername(updateInfo.Username) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.UnameNotFormat))
		return
	}
	if updateInfo.Sign != "" && !models.UserInfoUtil.CheckSign(updateInfo.Sign) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.SignNotFormat))
		return
	}
	if updateInfo.Sex != "" && updateInfo.Sex != models.Man && updateInfo.Sex != models.Woman {
		ctx.JSON(http.StatusOK, resp.Fail(definition.SexNotFormat))
		return
	}
	//	3. 进行修改
	err1 := models.UserInfoDao.UpdateUser(userClaim.Id, updateInfo)
	if err1 != nil && !errors.Is(err1, gorm.ErrRecordNotFound) {
		constants.MysqlErr(fmt.Sprintf("修改用户信息失败 : %v", updateInfo), err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success("信息修改成功"))
}

// GetUserInfo 获取用户基本信息
// api : /users/get/info  [get] LOGIN
func (h *UserHandler) GetUserInfo(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 获取用户信息
	queryUser, err1 := models.UserInfoDao.GetUserById(userClaim.Id)
	if err1 != nil && !errors.Is(err1, gorm.ErrRecordNotFound) {
		constants.MysqlErr("根据用户id获取用户信息异常", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(models.UserInfoUtil.TransToDtos(queryUser).First()))
}

// UploadUserCurPos 上传用户 当前经纬度坐标 -- 可以选择在用户登录,或进入程序时调用
// api : /users/upload/cur/pos  [post]
// post_args : {"x":"xxx","y":"xxx"} json LOGIN
func (h *UserHandler) UploadUserCurPos(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var posInfo = vo.UserVoHelper.NewUserVo().PositionVo
	if err := ctx.ShouldBind(&posInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 参数校验
	if ok := models.PointInfoUtil.CheckPointXY(posInfo.X, posInfo.Y); !ok {
		ctx.JSON(http.StatusOK, resp.Fail(definition.PosNotFormat))
		return
	}
	//	4. 检查其以前是否上传过
	exists, err1 := models.PointInfoDao.CheckPointIsExists(userClaim.Id)
	if err1 != nil && !errors.Is(err1, gorm.ErrRecordNotFound) {
		constants.MysqlErr("根据id查询位置信息失败", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	5. 存在更新，不存在创建
	var err2 error
	pointInfo := models.PointInfo{Id: userClaim.Id, Point: models.PointInfoUtil.MakePoint(posInfo.X, posInfo.Y)}
	if exists {
		err2 = models.PointInfoDao.UpdatePosById(pointInfo)
	} else {
		err2 = models.PointInfoDao.CreatePointInfo(pointInfo)
	}
	if err2 != nil {
		constants.MysqlErr(fmt.Sprintf("创建或更新 用户位置信息失败 , %v", pointInfo), err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success("用户位置信息已上传"))
}

const NearbyUserMAX = 50

// GetUserNearby 获取附近的用户 - 可以指定范围半径 (单位：km)  有最大限制 - DistanceMAX 500km - 每次最多50人
// api : /users/get/friend/nearby  [post]
// api_args : {"x":15,"y":30.5,"r":50,pagesize:1} json LOGIN
func (h *UserHandler) GetUserNearby(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := models.UserInfoUtil.IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var posXYR = vo.UserVoHelper.NewUserVo().PosXYR
	if err := ctx.ShouldBind(&posXYR); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 参数校验
	if posXYR.R <= 0 {
		ctx.JSON(http.StatusOK, resp.Fail(definition.RadiusTooSmall))
		return
	}
	if ok := models.PointInfoUtil.CheckPointXY(posXYR.X, posXYR.Y); !ok {
		ctx.JSON(http.StatusOK, resp.Fail(definition.XYNotFormat))
		return
	}
	if posXYR.Offset < 0 {
		posXYR.Offset = 0
	}
	//	4. 进行范围查询
	origin := models.PointInfoUtil.MakePoint(posXYR.X, posXYR.Y)
	iddisMap, err1 := models.PointInfoDao.GetUserByDistance(origin, posXYR.R*models.KM, NearbyUserMAX, posXYR.Offset)
	if err1 != nil && !errors.Is(err1, gorm.ErrRecordNotFound) {
		constants.MysqlErr("位置范围查询失败", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	for k := range iddisMap {
		if k == userClaim.Id {
			delete(iddisMap, k)
		}
	}
	if len(iddisMap) == 0 {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotFountAnyUser))
		return
	}
	//	5. 根据 id进行用户信息的查找
	infos, err2 := models.UserInfoDao.GetUsersByIds(iddisMap.Keys())
	if err2 != nil && !errors.Is(err2, gorm.ErrRecordNotFound) {
		constants.MysqlErr("根据id查询一些用户时出错", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	6. 进行最后的封装
	var dtos dto.NearbyUsers
	for _, info := range infos {
		dtos = append(dtos, dto.NearbyUserDto{
			Username: info.Username,
			Sex:      info.Sex,
			Sign:     info.Sign,
			Distance: iddisMap[info.Id],
		})
	}
	ctx.JSON(http.StatusOK, resp.Success(dtos))
}

// CreateGroup 创建群聊
// api : /users/create/group [post]
// post_args : {"group_name":"xxx","group_post":"xxx","max_person_num":xxx}  json LOGIN
func (h *UserHandler) CreateGroup(ctx *gin.Context) {
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
		LordId: userClaim.Id, AdminIds: "", MemberIds: "",
		CurPersonNum: 0, MaxPersonNum: createVo.MaxPersonNum,
	}
	err9 := models.GroupInfoDao.CreateGroup(groupInfo)
	if err9 != nil && !errors.Is(err9, gorm.ErrRecordNotFound) {
		constants.MysqlErr("添加群信息失败", err9)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success("创建成功"))
}
