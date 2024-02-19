package definition

import common "jcpd.cn/common/models"

var (
	ServerError       = common.NormalErr{Code: 1001, Msg: "服务器异常"}
	InvalidArgs       = common.NormalErr{Code: 1002, Msg: "请求参数有误"}
	ServerMaintaining = common.NormalErr{Code: 1003, Msg: "服务维护中，请稍后再试"}

	InvalidMobile = common.NormalErr{Code: 3001, Msg: "手机号格式不规范"}

	InvalidMode     = common.NormalErr{Code: 3011, Msg: "验证码模式错误"}
	FrequentCaptcha = common.NormalErr{Code: 3012, Msg: "验证码每分钟只能发送一次"}
	CaptchaNotSend  = common.NormalErr{Code: 3013, Msg: "请先发送验证码"}
	CaptchaError    = common.NormalErr{Code: 3014, Msg: "验证码错误"}

	PwdNotSame = common.NormalErr{Code: 3021, Msg: "两次密码不一致，请重新输入"}
	PwdError   = common.NormalErr{Code: 3022, Msg: "密码错误，请仔细检查"}

	UnameNotFormat = common.NormalErr{Code: 3031, Msg: "用户名不符合规范"}
	UnameExists    = common.NormalErr{Code: 3032, Msg: "该用户名已被使用"}
	UnameNotFound  = common.NormalErr{Code: 3033, Msg: "用户名不存在"}

	NotLogin = common.NormalErr{Code: 3010, Msg: "未登录或登录已过期"}
)
