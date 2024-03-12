package definition

import common "jcpd.cn/common/models"

var (
	ServerError       = common.NormalErr{Code: 1001, Msg: "服务器异常"}
	InvalidArgs       = common.NormalErr{Code: 1002, Msg: "请求参数有误"}
	ServerMaintaining = common.NormalErr{Code: 1003, Msg: "服务维护中，请稍后再试"}
	DataLoading       = common.NormalErr{Code: 1004, Msg: "数据加载中，请稍后再试"}

	NotLogin = common.NormalErr{Code: 3041, Msg: "未登录或登录已过期"}

	//	....

	PostTitleNotFormat = common.NormalErr{Code: 4001, Msg: "帖子标题不规范，应少于50字且不能为空"}
	PostTopicNotFormat = common.NormalErr{Code: 4002, Msg: "帖子主题不规范，应少于20字且不能为空"}
	PostBodyNotFormat  = common.NormalErr{Code: 4003, Msg: "帖子内容不规范，应少于1500字且不能为空"}
	PostNotFound       = common.NormalErr{Code: 4004, Msg: "帖子不存在或已被删除"}
	NotPostPublisher   = common.NormalErr{Code: 4005, Msg: "你不是该帖子的发布人"}

	PageNumNotFormat  = common.NormalErr{Code: 4011, Msg: "分页参数的页码有误，应为正整数"}
	PageSizeNotFormat = common.NormalErr{Code: 4012, Msg: "分页参数的每页大小有误，应为小于100的非负整数"}

	CommentNotFormat = common.NormalErr{Code: 4021, Msg: "帖子内容不符合规范"}
)
