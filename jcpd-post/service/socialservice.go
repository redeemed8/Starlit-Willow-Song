package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	common "jcpd.cn/common/models"
	"jcpd.cn/post/internal/constants"
	"jcpd.cn/post/internal/models"
	"jcpd.cn/post/internal/models/vo"
	"jcpd.cn/post/pkg/definition"
	"net/http"
	"strconv"
	"time"
)

// SocialHandler social路由的处理器 -- 用于管理各种接口的实现
type SocialHandler struct {
	cache definition.Cache
	errs  constants.Err_
}

func NewSocialHandler(type_ definition.CacheType) *SocialHandler {
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
	return &SocialHandler{cache: cache_}
}

const Like = "1"
const Dislike = "0"

// LikePost   点赞帖子
// api : /posts/social/like  [post]
// post_args : {"post_id":xxx}  json  LOGIN
func (h *SocialHandler) LikePost(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var likeVo = vo.SocialVoHelper.NewSocialVo().LikePostVo
	if err := ctx.ShouldBind(&likeVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 优先过滤掉部分帖子
	if exist := models.BloomFilters.Exist(likeVo.PostId); !exist {
		ctx.JSON(http.StatusOK, resp.Fail(definition.PostNotFound))
		return
	}
	//	4. 检查缓存中有无点赞信息  类型hash - key为 用户id，field为 帖子id，value为数字1，更快的知道用户点赞了哪些帖子
	PostIdStr := strconv.Itoa(int(likeVo.PostId))
	UserIdStr := strconv.Itoa(int(userClaim.Id))

	key := constants.PostLikePrefix + UserIdStr
	field := PostIdStr

	value0, err0 := h.cache.HGet(key, field)
	if h.errs.CheckRedisErr(err0) {
		constants.RedisErr("在缓存查询点赞信息失败", err0)
		//	TODO 限流，不然数据库压力大
	}
	if err0 != redis.Nil && value0 == Like {
		//	缓存中有点赞信息, 且值为 1 , 那就是他已经点过赞了, 不能重复的点赞
		ctx.JSON(http.StatusOK, resp.Success("点赞成功"))
		return
	}
	if err0 != redis.Nil && value0 == Dislike {
		//	缓存中有点赞信息, 且值为 0 , 那就是他没有点赞, 我们将其改为 1
		err1 := h.cache.HSet(key, field, Like, 5*time.Minute)
		if h.errs.CheckRedisErr(err1) {
			constants.RedisErr("更新缓存点赞状态信息失败", err1)
			//	TODO 限流，不然数据库压力大
		}
		ctx.JSON(http.StatusOK, resp.Success("点赞成功"))
		return
	}
	//	5. 缓存中没有信息，我们只能查一次数据库了

}

// DislikePost   取消点赞帖子
// api : /posts/social/dislike  [post]
// post_args : {"post_id":xxx}  json  LOGIN
func (h *SocialHandler) DislikePost(ctx *gin.Context) {

}
