package definition

import common "jcpd.cn/common/models"

var (
	ServerError = common.NormalErr{Code: 1001, Msg: "服务器异常"}
	InvalidArgs = common.NormalErr{Code: 1002, Msg: "请求参数有误"}

	InvalidMobile = common.NormalErr{Code: 3001, Msg: "手机号格式不规范"}
	InvalidMode   = common.NormalErr{Code: 3002, Msg: "验证码模式错误"}

	FrequentCaptcha = common.NormalErr{Code: 3003, Msg: "验证码每分钟只能发送一次"}
	PwdNotSame      = common.NormalErr{Code: 3004, Msg: "两次密码不一致，请重新输入"}

	UnameNotFormat = common.NormalErr{Code: 3005, Msg: "用户名不符合规范"}
	UnameExists    = common.NormalErr{Code: 3006, Msg: "该用户名已被使用"}

	CaptchaNotSend = common.NormalErr{Code: 3007, Msg: "请先发送验证码"}
	CaptchaError   = common.NormalErr{Code: 3008, Msg: "验证码错误"}

	NotLogin = common.NormalErr{Code: 3009, Msg: "未登录或登录已过期"}
)
