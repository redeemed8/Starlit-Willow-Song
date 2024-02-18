package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	common "jcpd.cn/common/models"
	"jcpd.cn/user/internal/constants"
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
// api : /users/getCaptcha?mobile=xxx&mode=xxx
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
		log.Println("Error : Redis exception ... ")
		//	TODO 降级处理，放入消息队列进行通知各个服务
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
