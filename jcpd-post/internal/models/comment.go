package models

import (
	"gorm.io/gorm"
	"jcpd.cn/post/internal/options"
	"jcpd.cn/post/utils"
	"time"
)

var CommentInfoDao commentInfoDao_
var CommentInfoUtil commentInfoUtil_

type commentInfoDao_ struct{ DB *gorm.DB }
type commentInfoUtil_ struct{}

func NewCommentInfoDao() {
	CommentInfoDao = commentInfoDao_{DB: options.C.DB}
}

// CommentInfo 帖子评论信息 - 联合索引
type CommentInfo struct {
	Id            uint32    `gorm:"primaryKey;autoIncrement"` //  主键 id,便于快速插入数据,虽然查询要回表，但好处是节省空间，同时防止页分裂
	CreatedAt     time.Time `gorm:"index:cp"`                 //  帖子创建时间
	PostId        uint32    `gorm:"not null;index:cp"`        //  所属帖子id
	PublisherId   uint32    `gorm:"not null"`                 //  发布人id
	PublisherName string    `gorm:"not null"`                 //  发布人用户名
	Body          string    `gorm:"not null"`                 //  评论内容
}

const CommentInfoTN = "7164_commentinfo"

// TableName 表名
func (comment *CommentInfo) TableName() string {
	return CommentInfoTN
}

// CreateTable 创建表
func (info *commentInfoDao_) CreateTable() {
	_ = info.DB.AutoMigrate(&CommentInfo{})
}

// CreateCommentInfo 创建一条评论
func (info *commentInfoDao_) CreateCommentInfo(comment *CommentInfo) error {
	return info.DB.Model(&CommentInfo{}).Create(comment).Error
}

// ------------------------------------

func (util *commentInfoUtil_) CheckContent(body string) bool {
	if body == "" || utils.CountCharacters(body) > 500 {
		return false
	}
	return true
}
