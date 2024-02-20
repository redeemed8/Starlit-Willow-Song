package definition

import common "jcpd.cn/common/models"

var (
	ServerError       = common.NormalErr{Code: 1001, Msg: "服务器异常"}
	InvalidArgs       = common.NormalErr{Code: 1002, Msg: "请求参数有误"}
	ServerMaintaining = common.NormalErr{Code: 1003, Msg: "服务维护中，请稍后再试"}

	InvalidMobile  = common.NormalErr{Code: 3001, Msg: "手机号格式不规范"}
	NotMatchMobile = common.NormalErr{Code: 3002, Msg: "用户名和手机号不匹配"}
	PhoneNotFound  = common.NormalErr{Code: 3003, Msg: "手机号信息不存在"}

	InvalidMode     = common.NormalErr{Code: 3011, Msg: "验证码模式错误"}
	FrequentCaptcha = common.NormalErr{Code: 3012, Msg: "验证码每分钟只能发送一次"}
	CaptchaNotSend  = common.NormalErr{Code: 3013, Msg: "请先发送验证码"}
	CaptchaError    = common.NormalErr{Code: 3014, Msg: "验证码错误"}

	PwdNotSame = common.NormalErr{Code: 3021, Msg: "两次密码不一致，请重新输入"}
	PwdError   = common.NormalErr{Code: 3022, Msg: "密码错误，请仔细检查"}

	UnameNotFormat = common.NormalErr{Code: 3031, Msg: "用户名不符合规范"}
	UnameExists    = common.NormalErr{Code: 3032, Msg: "该用户名已被使用"}
	UnameNotFound  = common.NormalErr{Code: 3033, Msg: "用户名不存在"}

	NotLogin      = common.NormalErr{Code: 3041, Msg: "未登录或登录已过期"}
	NotAuth2Token = common.NormalErr{Code: 3042, Msg: "无权限令牌不能修改"}
	Auth2TokenErr = common.NormalErr{Code: 3043, Msg: "令牌无效或已过期"}

	SignNotFormat = common.NormalErr{Code: 3051, Msg: "签名不符合规范"}
	SexNotFormat  = common.NormalErr{Code: 3052, Msg: "性别有误，未知性别"}

	PosNotFormat    = common.NormalErr{Code: 3061, Msg: "位置信息不合法"}
	RadiusTooSmall  = common.NormalErr{Code: 3062, Msg: "查询半径应该大于0km"}
	XYNotFormat     = common.NormalErr{Code: 3063, Msg: "经纬度信息不符合规范"}
	NotFountAnyUser = common.NormalErr{Code: 3064, Msg: "附近没有发现任何人哦"}
)
