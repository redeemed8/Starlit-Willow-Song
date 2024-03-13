package dto

import "time"

type CommentInfoDto struct {
	Id            uint32    `json:"id"`             //  主键 id
	CreatedAt     time.Time `json:"created_at"`     //  帖子创建时间
	PublisherName string    `json:"publisher_name"` //  发布人用户名
	Body          string    `json:"body"`           //  评论内容
	IsOwner       bool      `json:"is_owner"`       //	是否是评论的发布人
}

type CommentInfoDtos []CommentInfoDto
