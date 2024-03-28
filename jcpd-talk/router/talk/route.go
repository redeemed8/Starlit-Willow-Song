package talk

import (
	"github.com/gin-gonic/gin"
	"jcpd.cn/talk/internal/constants"
	"jcpd.cn/talk/pkg/definition"
	"jcpd.cn/talk/router"
	"jcpd.cn/talk/service"
	"log"
)

// 添加 talk路由到路由列表
func init() {
	log.Println(constants.Info("Application three api_init talk router ..."))
	router.Register(&RouterTalk{})
}

type RouterTalk struct {
}

// Router 实现方法，放置路由
func (*RouterTalk) Router(r *gin.Engine) {
	handler := service.NewTalkHandler(definition.CacheRedis)
	talkserviceGroup := r.Group("/talk")
	{
		talkserviceGroup.GET("/ws/connect/server", handler.ConnectServer)
		talkserviceGroup.GET("/ws/connect/friend", handler.TalkWithFriend)

		talkserviceGroup.GET("/message/unread/count", handler.GetUnReadMessageCount)
		talkserviceGroup.GET("/message/info/detail", handler.GetHistoryAndUnreadMsg)
	}
}
