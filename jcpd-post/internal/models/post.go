package models

import (
	"gorm.io/gorm"
	"jcpd.cn/post/internal/options"
)

var PostInfoDao PostInfoDao_
var PostInfoUtil PostInfoUtil_

type PostInfoDao_ struct{ DB *gorm.DB }
type PostInfoUtil_ struct{}

func NewPostInfoDao() {
	PostInfoDao = PostInfoDao_{DB: options.C.DB}
}

type PostInfo struct {
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
