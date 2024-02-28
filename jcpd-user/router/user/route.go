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
		userserviceGroup.POST("/register", handler.RegisterUser)

		userserviceGroup.POST("/login/mobile", handler.LoginMobile)
		userserviceGroup.POST("/login/passwd", handler.LoginPasswd)

		userserviceGroup.POST("/bind/mobile", handler.UserBindMobile)

		userserviceGroup.POST("/repasswd/check", handler.GetRepasswdToken)
		userserviceGroup.POST("/repasswd", handler.Repassword)

		userserviceGroup.POST("/update/info", handler.UpdateUserInfo)
		userserviceGroup.GET("/get/info", handler.GetUserInfo)
		userserviceGroup.GET("/search", handler.GetUserByName)

		userserviceGroup.POST("/upload/cur/pos", handler.UploadUserCurPos)
		userserviceGroup.POST("/get/friend/nearby", handler.GetUserNearby)
		userserviceGroup.GET("/friend/getlist", handler.GetOwnFriendList)
		userserviceGroup.POST("/delete/friend", handler.DeleteFriendById)

		userserviceGroup.POST("/group/create", handler.CreateGroup)
		userserviceGroup.GET("/group/getinfo/byid", handler.GetGroupInfoById)
		userserviceGroup.GET("/group/search", handler.GetGroupByName)
		userserviceGroup.POST("/group/update/info", handler.UpdateGroupInfo)
		userserviceGroup.GET("/group/getlist", handler.GetJoinedGroup)
		userserviceGroup.POST("/group/exit", handler.ExitGroup)
		userserviceGroup.POST("/group/set/admin", handler.ChooseUserToBeAdmin)

	}
}
