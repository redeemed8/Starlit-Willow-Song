package models

import (
	"gorm.io/gorm"
	"jcpd.cn/post/internal/options"
	"time"
)

var PostInfoDao postInfoDao_
var PostInfoUtil postInfoUtil_

type postInfoDao_ struct{ DB *gorm.DB }
type postInfoUtil_ struct{}

func NewPostInfoDao() {
	PostInfoDao = postInfoDao_{DB: options.C.DB}
}

// PostInfo 帖子类
type PostInfo struct {
	Id            uint32    `gorm:"primaryKey;autoIncrement"` //	主键 id -- 帖子id
	CreatedAt     time.Time //	帖子创建时间
	UpdatedAt     time.Time //	帖子的最近一次修改时间
	Title         string    `gorm:"not null;type:text"`         //	帖子标题
	TopicTag      string    `gorm:"not null;size:60;index:ttt"` //	主题标签
	Body          string    `gorm:"not null;type:text"`         //	帖子内容
	PublisherId   uint32    `gorm:"not null;index:ppp"`         //	发布人id
	PublisherName string    `gorm:"not null;size:31"`           //	发布人用户名
	Likes         int       `gorm:"default:0"`                  //	点赞数 - 热度
	AllowShow     bool      `gorm:"default:false"`              //	是否合法，是否可以展示出来
}

const PostInfoTN = "3491_postinfo"

// TableName 表名
func (post *PostInfo) TableName() string {
	return PostInfoTN
}

// CreateTable 创建表
func (info *postInfoDao_) CreateTable() {
	_ = info.DB.AutoMigrate(&PostInfo{})
}

// --------------------------------------------------

// CheckPostTitle 检查帖子标题
func (util *postInfoUtil_) CheckPostTitle(title string) bool {

	return false
}

// CheckPostTopicTag 检查帖子主题标签
func (util *postInfoUtil_) CheckPostTopicTag(topicTag string) bool {
	return false
}

// CheckPostBody 检查帖子内容
func (util *postInfoUtil_) CheckPostBody(body string) bool {
	return false
}

// CheckPostBase 检查帖子
func (util *postInfoUtil_) CheckPostBase(post PostInfo) bool {
	return false
}
