package group

import (
	"github.com/gin-gonic/gin"
	"jcpd.cn/user/pkg/definition"
	"jcpd.cn/user/router"
	"jcpd.cn/user/service"
	"log"
)

// 添加 group路由到路由列表
func init() {
	log.Println("Application one init group router ...")
	router.Register(&RouterGroup{})
}

type RouterGroup struct {
}

// Router 实现方法，放置路由
func (*RouterGroup) Router(r *gin.Engine) {
	handler := service.NewGroupHandler(definition.CacheRedis)
	applyserviceGroup := r.Group("/users/group")
	{
		applyserviceGroup.POST("/create", handler.CreateGroup)

		applyserviceGroup.GET("/getinfo/byid", handler.GetGroupInfoById)
		applyserviceGroup.GET("/search", handler.GetGroupByName)

		applyserviceGroup.POST("/update/info", handler.UpdateGroupInfo)
		applyserviceGroup.GET("/getlist", handler.GetJoinedGroup)

		applyserviceGroup.POST("/set/admin", handler.ChooseUserToBeAdmin)
		applyserviceGroup.POST("/cancel/admin", handler.CancelUserAdmin)

		applyserviceGroup.POST("/exit", handler.ExitGroup)
		applyserviceGroup.POST("/kick", handler.KickUserFromGroup)
	}
}
