package user

import (
	"github.com/gin-gonic/gin"
	"jcpd.cn/user/pkg/definition"
	"jcpd.cn/user/router"
	"jcpd.cn/user/service"
	"log"
)

// 添加 user路由到路由列表
func init() {
	log.Println("Application one init user router ...")
	router.Register(&RouterUser{})
}

type RouterUser struct {
}

// Router 实现方法，放置路由
func (*RouterUser) Router(r *gin.Engine) {
	handler := service.NewUserHandler(definition.CacheRedis)
	userserviceGroup := r.Group("/users")
	{
		userserviceGroup.GET("/getCaptcha", handler.GetCaptcha)
	}
}
