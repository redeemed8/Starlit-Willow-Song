package service

import (
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
	"jcpd.cn/user/internal/models/vo"
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

// GetCaptcha 根据手机号获取验证码
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
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerError))
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
			log.Printf("Failed to save the mobile and captcha to redis : %s : %s , cause by : %v \n", key, code, err1)
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
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerError))
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
	if err2 != nil {
		constants.MysqlErr("创建用户失败", err1)
		//	TODO 降级处理，放入消息队列进行通知各个服务，但这里不做，应在grpc里
		//	...
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerError))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success("注册成功，请登录"))
}

// LoginMobile 用户 手机号登录 -- 不用注册
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
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerError))
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
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerError))
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
			ctx.JSON(http.StatusOK, resp.Fail(definition.ServerError))
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
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerError))
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
	_ = models.UserInfoDao.UpdateUserByMap(userClaim.Id, columnMap)
	ctx.JSON(http.StatusOK, resp.Success("手机号已更新"))
}
