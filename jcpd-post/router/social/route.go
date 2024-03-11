package social

import (
	"github.com/gin-gonic/gin"
	"jcpd.cn/post/internal/constants"
	"jcpd.cn/post/pkg/definition"
	"jcpd.cn/post/router"
	"jcpd.cn/post/service"
	"log"
)

// 添加 social 路由到路由列表
func init() {
	log.Println(constants.Info("Application two api_init social router ..."))
	router.Register(&RouterSocial{})
}

type RouterSocial struct {
}

// Router 实现方法，放置路由
func (*RouterSocial) Router(r *gin.Engine) {
	handler := service.NewSocialHandler(definition.CacheRedis)
	postserviceGroup := r.Group("/posts/social")
	{
		postserviceGroup.POST("/like", handler.LikePost)
		postserviceGroup.POST("/dislike", handler.DislikePost)

		postserviceGroup.POST("/comment/publish", handler.PublishComment)
		postserviceGroup.POST("/comment/delete", handler.DeleteComment)
		postserviceGroup.GET("/comment/getnew", handler.GetNewestComment)
	}
}
