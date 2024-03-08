package main

import (
	"github.com/gin-gonic/gin"
	"jcpd.cn/common"
	"jcpd.cn/user/api/api_init"
	_ "jcpd.cn/user/internal/init"
	"jcpd.cn/user/internal/options"
	"jcpd.cn/user/router"
)

func main() {
	//	默认配置的 Gin 引擎实例
	r := gin.Default()
	//  据路由列表 初始化路由引擎
	router.InitRouter(r)
	//	grpc 服务注册
	grpc := api_init.RegisterGrpc()
	//	开启定时任务
	go router.TimerTasks.Start()
	//	启动服务
	common.Run(r, options.C.App.Server.Port, options.C.App.Server.Name, grpc.Stop, router.TimerTasks.Check)
}
