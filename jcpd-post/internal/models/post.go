package models

import (
	"gorm.io/gorm"
	"jcpd.cn/post/internal/options"
	"time"
)

var PostInfoDao PostInfoDao_
var PostInfoUtil PostInfoUtil_

type PostInfoDao_ struct{ DB *gorm.DB }
type PostInfoUtil_ struct{}

func NewPostInfoDao() {
	PostInfoDao = PostInfoDao_{DB: options.C.DB}
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
func (info *PostInfoDao_) CreateTable() {
	_ = info.DB.AutoMigrate(&PostInfo{})
}

// --------------------------------------------------
