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
	if likeVo.PostId < 1 { //	过滤器接收的最小id为1
		ctx.JSON(http.StatusOK, resp.Fail(definition.PostNotFound))
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
	queryLike, err2 := models.LikeInfoDao.GetLikeByTwoId(userClaim.Id, likeVo.PostId)
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("根据用户id和帖子id获取点赞信息出错", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryLike.Id != 0 {
		//	id不为0代表存在点赞记录, 不能重复点赞
		ctx.JSON(http.StatusOK, resp.Success("点赞成功"))
		return
	}
	//	6. 数据库中没有点赞记录, 我们在 redis中为其创建点赞记录
	err3 := h.cache.HSet(key, field, Like, 30*time.Minute)
	if h.errs.CheckRedisErr(err3) {
		constants.RedisErr("在缓存中创建点赞记录出错", err3)
		//	TODO 限流啦 ...

		//	redis 无法存储，只能用 mysql暂时顶替
		err4 := models.LikeInfoDao.CreateLikeInfo(models.LikeInfo{PostId: likeVo.PostId, UserId: userClaim.Id})
		if h.errs.CheckMysqlErr(err4) {
			constants.MysqlErr("在mysql创建点赞记录失败", err4)
			ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
			return
		}
	}
	//	7. 成功
	ctx.JSON(http.StatusOK, resp.Success("点赞成功"))
}

// DislikePost   取消点赞帖子
// api : /posts/social/dislike  [post]
// post_args : {"post_id":xxx}  json  LOGIN
func (h *SocialHandler) DislikePost(ctx *gin.Context) {
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
	if likeVo.PostId < 1 { //	过滤器接收的最小id为1
		ctx.JSON(http.StatusOK, resp.Fail(definition.PostNotFound))
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
	if err0 != redis.Nil && value0 == Dislike {
		//	缓存中有点赞信息, 且值为 0 , 那就是他已取消点赞, 直接返回即可
		ctx.JSON(http.StatusOK, resp.Success("点赞已取消"))
		return
	}
	if err0 != redis.Nil && value0 == Like {
		//	缓存中有点赞信息, 且值为 1 , 那就是他已点赞, 我们将其改为 0
		err1 := h.cache.HSet(key, field, Dislike, 5*time.Minute)
		if h.errs.CheckRedisErr(err1) {
			constants.RedisErr("更新缓存点赞状态信息失败", err1)
			//	TODO 限流，不然数据库压力大
		}
		ctx.JSON(http.StatusOK, resp.Success("点赞已取消"))
		return
	}
	//	5. 缓存中没有信息，我们只能查一次数据库了
	queryLike, err2 := models.LikeInfoDao.GetLikeByTwoId(userClaim.Id, likeVo.PostId)
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("根据用户id和帖子id获取点赞信息出错", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryLike.Id == 0 {
		//	id为0代表不存在点赞记录, 直接返回即可
		ctx.JSON(http.StatusOK, resp.Success("点赞已取消"))
		return
	}
	//	6. 数据库中有点赞记录, 我们在 redis中为其点赞记录 记为0
	err3 := h.cache.HSet(key, field, Dislike, 30*time.Minute)
	if h.errs.CheckRedisErr(err3) {
		constants.RedisErr("在缓存中创建点赞记录出错", err3)
		//	TODO 限流啦 ...

		//	redis 无法存储，只能暂时在 mysql删除
		err4 := models.LikeInfoDao.DeleteLikeByTwoId(userClaim.Id, likeVo.PostId)
		if h.errs.CheckMysqlErr(err4) {
			constants.MysqlErr("在mysql删除点赞记录失败", err4)
			ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
			return
		}
	}
	//	7. 成功
	ctx.JSON(http.StatusOK, resp.Success("点赞已取消"))
}

//	明天写评论
