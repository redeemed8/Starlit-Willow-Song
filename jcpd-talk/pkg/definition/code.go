package definition

import common "jcpd.cn/common/models"

var (
	ServerError       = common.NormalErr{Code: 1001, Msg: "服务器异常"}
	InvalidArgs       = common.NormalErr{Code: 1002, Msg: "请求参数有误"}
	ServerMaintaining = common.NormalErr{Code: 1003, Msg: "服务维护中，请稍后再试"}
	DataLoading       = common.NormalErr{Code: 1004, Msg: "数据加载中，请稍后再试"}

	UserNotFound = common.NormalErr{Code: 3034, Msg: "未找到相关的用户信息"}
	NotLogin     = common.NormalErr{Code: 3041, Msg: "未登录或登录已过期"}

	//	....

	NotFriend = common.NormalErr{Code: 5001, Msg: "你们还不是好友呢，请先添加为好友"}
)
