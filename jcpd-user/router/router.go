package router

import (
	"github.com/gin-gonic/gin"
	"jcpd.cn/user/internal/models"
	"jcpd.cn/user/service"
)

// Router 路由接口
type Router interface {
	Router(r *gin.Engine)
}

// 路由列表
var routers []Router

// InitRouter 根据路由列表 初始化路由引擎
func InitRouter(r *gin.Engine) {
	//	创建表 和 初始化
	models.InitUser()
	//  根据路由列表 初始化路由引擎
	for _, ro := range routers {
		ro.Router(r)
	}
	//	开启定时任务
	service.TimerTasks.Start()
}

// Register	用于注册路由
func Register(ro ...Router) {
	routers = append(routers, ro...)
}
