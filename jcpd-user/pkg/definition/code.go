package definition

import common "jcpd.cn/common/models"

var (
	InvalidMobile   = common.NormalErr{Code: 3001, Msg: "手机号格式不规范"}
	InvalidMode     = common.NormalErr{Code: 3002, Msg: "验证码模式错误"}
	FrequentCaptcha = common.NormalErr{Code: 3003, Msg: "验证码每分钟只能发送一次"}
)
