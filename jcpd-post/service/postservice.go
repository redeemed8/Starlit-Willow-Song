package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	common "jcpd.cn/common/models"
	commonJWT "jcpd.cn/common/utils/jwt"
	"jcpd.cn/post/internal/constants"
	"jcpd.cn/post/internal/models"
	"jcpd.cn/post/internal/models/vo"
	"jcpd.cn/post/pkg/definition"
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
// post
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
	m := map[string]string{"post_id": strconv.Itoa(int(postInfo.Id)), "ret_msg": "帖子已提交，等待审核通过后即可发布"}
	ctx.JSON(200, resp.Success(m))
}

// GetPostSummaryHot 获取帖子简介 - 指定 页码(最小页码为1) 每页数量(<50) - 优先点赞热度排序 + redis缓存id
// api : /posts/get/summary/hot?pagenum=xxx&size=xxx  [get]
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
	//	7. 如果缓存没有命中，只能查数据库了
	postInfos, err99 = models.PostInfoDao.SimpleGetPostsPage(page)
	if h.errs.CheckMysqlErr(err99) {
		constants.MysqlErr("分页查询帖子信息出错", err99)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	//	8. 将查到的记录 添加到 redis
	err8 := h.cache.Put(constants.HotPostSummary, postInfos.ToIdStr(), 90*time.Minute)
	if h.errs.CheckRedisErr(err8) {
		constants.RedisErr("获取redis缓存帖子id出错", err8)
		//	TODO  此处 还应该进行 服务降级处理 -- 减少访问到达量
	}
	//	9 返回帖子简述
	ctx.JSON(http.StatusOK, resp.Success(postInfos.ToDtos()))
}

// GetPostSummaryTime 获取帖子简介 - 指定 每页数量 以及 上次分页中的的最小id - 优先发布时间排序
// api : /posts/get/summary/time?size=xxx&lmid=  [get]
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

// GetPostDetails 根据id ，获取帖子详细内容
// api : /posts/get/detail?postid=xxx
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
	//	3. 查数据库 id
	queryPost, err2 := models.PostInfoDao.GetPostById(postId)
	if h.errs.CheckMysqlErr(err2) {
		constants.MysqlErr("根据id获取活动信息失败", err2)
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerMaintaining))
		return
	}
	if queryPost.PublisherName == "" {
		ctx.JSON(http.StatusOK, resp.Fail(definition.PostNotFound))
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(queryPost.Body))
}

//	点赞 ... 另起一个文件
