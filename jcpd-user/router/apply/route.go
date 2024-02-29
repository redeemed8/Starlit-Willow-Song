package apply

import (
	"github.com/gin-gonic/gin"
	"jcpd.cn/user/pkg/definition"
	"jcpd.cn/user/router"
	"jcpd.cn/user/service"
	"log"
)

// 添加 apply路由到路由列表
func init() {
	log.Println("Application one init apply router ...")
	router.Register(&RouterApply{})
}

type RouterApply struct {
}

// Router 实现方法，放置路由
func (*RouterApply) Router(r *gin.Engine) {
	handler := service.NewApplyHandler(definition.CacheRedis)
	applyserviceGroup := r.Group("/users/apply")
	{
		applyserviceGroup.POST("/tobe/friend", handler.ApplyToBeFriend)
		applyserviceGroup.GET("/get/all", handler.GetAllAppliesByStatus)
		applyserviceGroup.POST("/update/status", handler.UpdateApplyStatus)
		applyserviceGroup.POST("/toadd/group", handler.ApplyToGroup)
		applyserviceGroup.POST("/update/group-status", handler.UpdateApplyGroupStatus)
	}
}
