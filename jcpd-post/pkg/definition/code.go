package definition

import common "jcpd.cn/common/models"

var (
	ServerError       = common.NormalErr{Code: 1001, Msg: "服务器异常"}
	InvalidArgs       = common.NormalErr{Code: 1002, Msg: "请求参数有误"}
	ServerMaintaining = common.NormalErr{Code: 1003, Msg: "服务维护中，请稍后再试"}
)
