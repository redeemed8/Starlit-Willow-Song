package main

import (
	"github.com/gin-gonic/gin"
	"jcpd.cn/common"
	_ "jcpd.cn/post/internal/init"
	"jcpd.cn/post/internal/options"
	"jcpd.cn/post/router"
)

func main() {
	//	默认配置的 Gin 引擎实例
	r := gin.Default()
	//  据路由列表 初始化路由引擎
	router.InitRouter(r)
	//	开启定时任务
	go router.TimerTasks.Start()
	//	启动服务
	common.Run(r, options.C.App.Server.Port, options.C.App.Server.Name, router.TimerTasks.Check)
}
