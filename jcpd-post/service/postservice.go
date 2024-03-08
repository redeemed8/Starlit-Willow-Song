package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	common "jcpd.cn/common/models"
	commonJWT "jcpd.cn/common/utils/jwt"
	"jcpd.cn/post/internal/constants"
	"jcpd.cn/post/internal/models"
	"jcpd.cn/post/internal/models/vo"
	"jcpd.cn/post/pkg/definition"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// PostHandler post路由的处理器 -- 用于管理各种接口的实现
type PostHandler struct {
	cache definition.Cache
	errs  constants.Err_
}

func NewPostHandler(type_ definition.CacheType) *PostHandler {
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
	return &PostHandler{cache: cache_}
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

// Publish 发布一篇帖子
// api : /posts/publish  [post]
// post_args : {"title":"xxx","topic":"xxx","body":"xxx"}  json LOGIN
func (h *PostHandler) Publish(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var publishVo = vo.PostVoHelper.NewPostVo().PublishPostVo
	if err := ctx.ShouldBind(&publishVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 参数校验
	postInfo := models.PostInfo{Title: publishVo.Title, TopicTag: publishVo.TopicTag, Body: publishVo.Body}
	if err := models.PostInfoUtil.CheckPostBase(postInfo); err != nil {
		ctx.JSON(http.StatusOK, resp.Fail(*err))
		return
	}
	//	4. 创建帖子信息
	postInfo.TopicTag = "#" + postInfo.TopicTag
	postInfo.PublisherId = userClaim.Id
	postInfo.PublisherName = userClaim.Username
	if err := models.PostInfoDao.CreatePost(&postInfo); h.errs.CheckMysqlErr(err) {
		constants.MysqlErr("创建帖子postinfo信息出错", err)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	id := strconv.Itoa(int(postInfo.Id))
	m := map[string]string{"post_id": id, "ret_msg": "帖子已提交，等待审核通过后即可发布"}
	//	5. 添加到布隆过滤器中
	models.BloomFilters.Add(postInfo.Id)
	ctx.JSON(200, resp.Success(m))
}

// GetPostSummaryHot 获取帖子简介 - 指定 页码(最小页码为1) 每页数量(<50) - 优先点赞热度排序 + redis缓存id
// api : /posts/get/summary/hot?pagenum=xxx&size=xxx  [get]  LOGIN
func (h *PostHandler) GetPostSummaryHot(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, _ := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 获取路径参数
	pagenum := ctx.Query("pagenum")
	pagesize := ctx.Query("size")
	//	3. 校验路径参数
	page, err1 := models.PostInfoUtil.CheckPage(pagenum, pagesize)
	if err1 != nil {
		ctx.JSON(http.StatusOK, resp.Fail(*err1))
		return
	}
	var postInfos = make(models.PostInfos, 0)
	var err99 error
	//	4. 判断是否查询的是热点页 ， 也就是第一页
	if page.PageNum > 1 {
		//	不是热点页，没有缓存，直接查询。
		postInfos, err99 = models.PostInfoDao.SimpleGetPostsPage(page)
		if h.errs.CheckMysqlErr(err99) {
			constants.MysqlErr("分页查询帖子信息出错", err99)
			ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
			return
		}
		ctx.JSON(http.StatusOK, resp.Success(postInfos.ToDtos()))
		return
	}
	//	5. 如果查询的是热点页，先查 redis的缓存
	idsStr, err2 := h.cache.Get(constants.HotPostSummary)
	if h.errs.CheckRedisErr(err2) {
		//  说明查询缓存出错，有可能是redis宕机了，此处开启异常处理，但不结束流程
		constants.RedisErr("获取redis缓存帖子id出错", err2)
		//	TODO  此处 还应该进行 服务降级处理 -- 减少访问到达量
	}
	//	6. 如果缓存命中, 简单查询后返回
	if err2 == nil && idsStr != "" {
		postInfos, err99 = models.PostInfoDao.GetPostsInIds(idsStr)
		if h.errs.CheckMysqlErr(err99) {
			constants.MysqlErr("分页查询帖子信息出错", err99)
			ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
			return
		}
		ctx.JSON(http.StatusOK, resp.Success(postInfos.ToDtos()))
		return
	}
	//	7. 如果缓存没有命中，只能 尝试获取分布式锁
	lockKey := constants.HotPostSummaryLockPrefix + pagenum
	err7 := h.cache.SetNX(lockKey, "1")
	if h.errs.CheckRedisErr(err7) {
		constants.RedisErr("创建post内容缓存时，获取分布式锁失败", err7)
		//	 TODO 要是redis宕机就只能限流了。。。
	}
	if err7 == redis.Nil {
		ctx.JSON(http.StatusOK, resp.Fail(definition.DataLoading)) //	没抢到锁，先返回一会再刷新重试
		return
	}
	//	8. 然后查数据库
	postInfos, err99 = models.PostInfoDao.SimpleGetPostsPage(page)
	if h.errs.CheckMysqlErr(err99) {
		constants.MysqlErr("分页查询帖子信息出错", err99)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	9. 将查到的记录 添加到 redis
	err8 := h.cache.Put(constants.HotPostSummary, postInfos.ToIdStr(), 90*time.Minute)
	if h.errs.CheckRedisErr(err8) {
		constants.RedisErr("获取redis缓存帖子id出错", err8)
		//	TODO  此处 还应该进行 服务降级处理 -- 减少访问到达量
	}
	//	10. 释放分布式锁, 返回帖子简述
	_ = h.cache.Delete(lockKey)
	ctx.JSON(http.StatusOK, resp.Success(postInfos.ToDtos()))
}

// GetPostSummaryTime 获取帖子简介 - 指定 每页数量 以及 上次分页中的的最小id - 优先发布时间排序
// api : /posts/get/summary/time?size=xxx&lmid=  [get]  LOGIN
// args_explain : 如果是第一次分页查询，参数传空即可，如果lmid参数错误，将默认为不开启优化
func (h *PostHandler) GetPostSummaryTime(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, _ := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 获取路径参数
	pagesize := ctx.Query("size")
	lmid := ctx.Query("lmid")
	//	3. 校验路径参数
	page, err := models.PostInfoUtil.CheckPage("10", pagesize)
	if err != nil {
		ctx.JSON(http.StatusOK, resp.Fail(*err))
		return
	}
	lastMinPostId, ok := models.PostInfoUtil.CheckLmid(lmid)
	//	4. 分页查询
	infos, err2 := models.PostInfoDao.SeniorGetPostPage(page, lastMinPostId, ok)
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("优化分页查询帖子信息出错", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(infos.ToDtos()))
}

// GetPostDetails 根据id ，获取帖子详细内容 + redis 缓存
// api : /posts/get/detail?postid=xxx  [get]  LOGIN
func (h *PostHandler) GetPostDetails(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, _ := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 获取路径
	postIdStr := ctx.Query("postid")
	postId, err1 := models.PostInfoUtil.CheckPostIdStr(postIdStr)
	if err1 != nil {
		ctx.JSON(http.StatusOK, resp.Fail(*err1))
		return
	}
	//	3. 访问布隆过滤器，不存在则一定不存在，防止缓存穿透
	if exist := models.BloomFilters.Exist(postId); !exist {
		ctx.JSON(http.StatusOK, resp.Fail(definition.PostNotFound))
		return
	}
	//	4. 访问redis缓存
	cachePostBody, err2 := h.cache.Get(constants.PostBodyPrefix + postIdStr)
	if h.errs.CheckRedisErr(err2) {
		constants.RedisErr("获取帖子内容缓存出错,", err2)
		//	TODO 降级处理
	}
	if cachePostBody == models.SpecialSymbol { //	相当于 空值 nil
		ctx.JSON(http.StatusOK, resp.Fail(definition.PostNotFound))
		return
	}
	if err2 == nil && cachePostBody != "" {
		//	缓存命中, 直接返回
		ctx.JSON(http.StatusOK, resp.Success(cachePostBody))
		return
	}
	//	5. 缓存未命中 ,查数据库 id，先尝试获取分布式锁，防止因并发过大导致重复的缓存建立
	lockKey := constants.PostBodyLockPrefix + postIdStr
	err7 := h.cache.SetNX(lockKey, "1")
	if h.errs.CheckRedisErr(err7) {
		constants.RedisErr("创建post内容缓存时，获取分布式锁失败", err7)
		//	 TODO 要是redis宕机就只能限流了。。。
	}
	if err7 == redis.Nil {
		ctx.JSON(http.StatusOK, resp.Fail(definition.DataLoading)) //	没抢到锁，先返回一会再刷新重试
		return
	}
	//	6. 抢到锁了，进行数据库查询
	queryPost, err3 := models.PostInfoDao.GetPostById(postId)
	if h.errs.CheckMysqlErr(err3) {
		constants.MysqlErr("根据id获取活动信息失败", err3)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryPost.PublisherName == "" {
		//	7.如果数据库中也不存在，说明布隆过滤器出错，所以我们要在redis中存一个标记
		err8 := h.cache.Put(constants.PostBodyPrefix+postIdStr, models.SpecialSymbol, 10*time.Minute)
		if err8 != nil {
			constants.RedisErr("为不存在的id设置redis缓存失败", err8)
		}
		ctx.JSON(http.StatusOK, resp.Fail(definition.PostNotFound))
		return
	}
	//	8. 数据库中存在，建立redis缓存
	cacheTime := time.Duration(30 + rand.Intn(11)) //	设置随机数，防止同时失效导致缓存雪崩
	err4 := h.cache.Put(constants.PostBodyPrefix+postIdStr, queryPost.Body, cacheTime*time.Minute)
	if err4 != nil {
		constants.RedisErr("为帖子内容设置redis缓存失败", err4)
		//	TODO 降级处理
	}
	//	9. 释放分布式锁
	_ = h.cache.Delete(lockKey)
	ctx.JSON(http.StatusOK, resp.Success(queryPost.Body))
}

// UpdatePost 修改帖子信息, 不想修改参数的不用传递
// api : /posts/updates/info  [post]
// post_args : {"post_id":xxx,"title":"xxx","topic":"xxx","body":"xxx"}  json LOGIN
func (h *PostHandler) UpdatePost(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var updateVo = vo.PostVoHelper.NewPostVo().UpdatePostVo
	if err := ctx.ShouldBind(&updateVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	//	3. 参数校验
	if updateVo.PostId == 0 {
		ctx.JSON(http.StatusOK, resp.Fail(definition.PostNotFound))
		return
	}
	postinfo, err1 := models.PostInfoUtil.MakePostInfo(updateVo.Title, updateVo.TopicTag, updateVo.Body)
	if err1 != nil {
		ctx.JSON(http.StatusOK, resp.Success(*err1))
		return
	}
	//	4. 校验身份 - 是不是帖子的发布人
	ownerId, err0 := models.PostInfoDao.GetPostOwnerById(updateVo.PostId)
	if h.errs.CheckMysqlErr(err0) {
		constants.MysqlErr("根据id获取发布人id出错", err0)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if ownerId != userClaim.Id {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotPostPublisher))
		return
	}
	//	5. 为了保证最终一致性，我们用先更新数据库，然后删除缓存
	err2 := models.PostInfoDao.UpdatePostByInfo(updateVo.PostId, postinfo)
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("根据id更新帖子信息出错", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if errors.Is(err2, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.PostNotFound))
		return
	}
	//	5. 删除缓存, 这里不是顺便更新缓存的原因是：这个帖子不一定有人访问，不一定需要缓存，也能很大程度上节省 redis空间
	err3 := h.cache.Delete(constants.PostBodyPrefix + strconv.Itoa(int(updateVo.PostId)))
	if h.errs.CheckRedisErr(err3) {
		constants.RedisErr("删除redis缓存帖子信息出错", err3)
	}
	ctx.JSON(http.StatusOK, resp.Success("帖子信息已更新"))
}

// DeletePost 	 通过帖子id 删除帖子
// api : /posts/delete/one  [post]
// post_args : {"post_id":xxx}  json LOGIN
func (h *PostHandler) DeletePost(ctx *gin.Context) {
	resp := common.NewResp()
	//	1. 校验登录
	normalErr, userClaim := IsLogin(ctx, resp)
	if normalErr != nil {
		return
	}
	//	2. 绑定参数
	var deleteVo = vo.PostVoHelper.NewPostVo().DeletePostVo
	if err := ctx.ShouldBind(&deleteVo); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Fail(definition.InvalidArgs))
		return
	}
	if deleteVo.PostId == 0 {
		ctx.JSON(http.StatusOK, resp.Fail(definition.PostNotFound))
		return
	}
	//	3. 校验身份 - 是不是帖子的发布人
	ownerId, err0 := models.PostInfoDao.GetPostOwnerById(deleteVo.PostId)
	if h.errs.CheckMysqlErr(err0) {
		constants.MysqlErr("根据id获取发布人id出错", err0)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if ownerId != userClaim.Id {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotPostPublisher))
		return
	}
	//	4. 这里我们采用先删除缓存，再删除数据库，再删除缓存的方式 ..
	//	先删除缓存，如果报错，直接返回删除失败，避免了先删除数据库，然后缓存删除失败的情况
	err1 := h.cache.Delete(constants.PostBodyPrefix + strconv.Itoa(int(deleteVo.PostId)))
	if h.errs.CheckRedisErr(err1) {
		constants.RedisErr("删除redis缓存帖子信息出错", err1)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	再删除数据库
	err2 := models.PostInfoDao.DeletePostById(deleteVo.PostId)
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("删除数据库帖子信息出错", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	再次删除缓存，防止在删除数据库时有人访问并添加了缓存
	err3 := h.cache.Delete(constants.PostBodyPrefix + strconv.Itoa(int(deleteVo.PostId)))
	if h.errs.CheckRedisErr(err3) {
		constants.RedisErr("删除redis缓存帖子信息出错", err3)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success("帖子删除成功"))
}

// GetNotReviewedPost  获取到所有未审核的帖子 - 可以做一部分管理员用户, 仅限管理员使用
// api : /posts/getinfo/not-reviewed  [get]  LOGIN
func (h *PostHandler) GetNotReviewedPost(ctx *gin.Context) {

}

// ReviewPost   审核帖子信息 - 可以做一部分管理员用户, 仅限管理员使用
// api : /posts/review  [post]
// post_args : {"post_id":xxx,"cur_status":"xxx","to_status":"xxx"}  json LOGIN
func (h *PostHandler) ReviewPost(ctx *gin.Context) {

}
