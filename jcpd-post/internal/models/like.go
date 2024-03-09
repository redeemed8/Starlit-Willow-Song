package models

import (
	"gorm.io/gorm"
	"jcpd.cn/post/internal/options"
)

var LikeInfoDao likeInfoDao_
var LikeInfoUtil likeInfoUtil_

type likeInfoDao_ struct{ DB *gorm.DB }
type likeInfoUtil_ struct{}

func NewLikeInfoDao() {
	LikeInfoDao = likeInfoDao_{DB: options.C.DB}
}

// LikeInfo 帖子点赞信息 - 联合索引
type LikeInfo struct {
	Id     int    `gorm:"primaryKey;autoIncrement"` //	主键 id
	UserId uint32 `gorm:"not null;index:likeid"`    //	用户 id
	PostId uint32 `gorm:"not null;index:likeid'"`   //  帖子 id
}

const LikeInfoTN = "6481_likeinfo"

// TableName 表名
func (like *LikeInfo) TableName() string {
	return LikeInfoTN
}

// CreateTable 创建表
func (info *likeInfoDao_) CreateTable() {
	_ = info.DB.AutoMigrate(&LikeInfo{})
}

// CreateLikeInfo 创建一个点赞信息
func (info *likeInfoDao_) CreateLikeInfo(like LikeInfo) error {
	return info.DB.Model(&LikeInfo{}).Create(like).Error
}

// GetLikeByTwoId 根据用户id 和 帖子id 获取点赞信息
func (info *likeInfoDao_) GetLikeByTwoId(userId uint32, postId uint32) (LikeInfo, error) {
	var like LikeInfo
	result := info.DB.Model(&LikeInfo{}).Where("user_id = ? and post_id = ?", userId, postId).First(&like)
	return like, result.Error
}

func (info *likeInfoDao_) DeleteLikeByTwoId(userId uint32, postId uint32) error {
	return info.DB.Model(&LikeInfo{}).Where("user_id = ? and post_id = ?", userId, postId).Delete(&LikeInfo{}).Error
}

// -----------------------------------------
