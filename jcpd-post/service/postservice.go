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
