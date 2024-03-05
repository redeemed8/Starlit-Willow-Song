package definition

import common "jcpd.cn/common/models"

var (
	ServerError       = common.NormalErr{Code: 1001, Msg: "服务器异常"}
	InvalidArgs       = common.NormalErr{Code: 1002, Msg: "请求参数有误"}
	ServerMaintaining = common.NormalErr{Code: 1003, Msg: "服务维护中，请稍后再试"}
	NotLogin          = common.NormalErr{Code: 3041, Msg: "未登录或登录已过期"}

	//	....

	PostTitleNotFormat = common.NormalErr{Code: 4001, Msg: "帖子标题不规范，应少于50字且不能为空"}
	PostTopicNotFormat = common.NormalErr{Code: 4002, Msg: "帖子主题不规范，应少于20字且不能为空"}
	PostBodyNotFormat  = common.NormalErr{Code: 4003, Msg: "帖子内容不规范，应少于1500字且不能为空"}
)
