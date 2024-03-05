package dto

import "time"

type PostInfoDto struct {
	Id            uint32    `json:"id"`             //	主键 id -- 帖子id
	Title         string    `json:"title"`          //	帖子标题
	TopicTag      string    `json:"topic_tag"`      //	主题标签
	Body          string    `json:"body"`           //	帖子内容
	PublisherName string    `json:"publisher_name"` //	发布人用户名
	PublishTime   time.Time `json:"publish_time"`   //	帖子发布时间
	Likes         int       `json:"likes"`          //	点赞数 - 热度
	ReviewStatus  string    `json:"review_status"`  //	审核状态, 0-未审核，1-已通过，2-已驳回
	Reason        string    `json:"reason"`         //	驳回原因 -- 保存3天
}
