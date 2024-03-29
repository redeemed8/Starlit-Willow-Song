package post

import (
	"github.com/gin-gonic/gin"
	"jcpd.cn/post/internal/constants"
	"jcpd.cn/post/pkg/definition"
	"jcpd.cn/post/router"
	"jcpd.cn/post/service"
	"log"
)

// 添加 post路由到路由列表
func init() {
	log.Println(constants.Info("Application two api_init post router ..."))
	router.Register(&RouterPost{})
}

type RouterPost struct {
}

// Router 实现方法，放置路由
func (*RouterPost) Router(r *gin.Engine) {
	handler := service.NewPostHandler(definition.CacheRedis)
	postserviceGroup := r.Group("/posts")
	{
		postserviceGroup.POST("/publish", handler.Publish)
		postserviceGroup.GET("/get/summary/hot", handler.GetPostSummaryHot)
		postserviceGroup.GET("/get/summary/time", handler.GetPostSummaryTime)
		postserviceGroup.GET("/get/detail", handler.GetPostDetails)
		postserviceGroup.POST("/updates/info", handler.UpdatePost)
		postserviceGroup.POST("/delete/one", handler.DeletePost)
		postserviceGroup.GET("/getinfo/not-reviewed", handler.GetNotReviewedPost)
		postserviceGroup.POST("/review", handler.ReviewPost)
	}
}
