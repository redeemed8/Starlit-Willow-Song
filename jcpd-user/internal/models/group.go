package models

import (
	"gorm.io/gorm"
	"jcpd.cn/user/internal/options"
	"time"
)

var GroupInfoDao GroupInfoDao_
var GroupInfoUtil GroupInfoUtil_

type GroupInfoDao_ struct{ DB *gorm.DB }
type GroupInfoUtil_ struct{}

func NewGroupInfoDao() {
	GroupInfoDao = GroupInfoDao_{DB: options.C.DB}
}

type GroupInfo struct {
	Id           uint32    `gorm:"primaryKey;autoIncrement"` //	主键 id
	GroupName    string    `gorm:"not null"`                 //	群名称
	GroupPost    string    `gorm:"not null"`                 //	群公告
	LordId       uint32    `gorm:"not null;index"`           //	群主 id
	AdminIds     string    `gorm:"type:longtext"`            //	管理员 id
	MemberIds    string    `gorm:"type:longtext"`            //	成员 id
	CurPersonNum int       `gorm:"default:0"`                //	当前人数
	MaxPersonNum int       `gorm:"not null;default:100"`     //	最大人数
	CreatedAt    time.Time //	创建时间
	UpdatedAt    time.Time //	更新时间
}

const GroupInfoTN = "5433_group"

// TableName 表名
func (table *GroupInfo) TableName() string {
	return GroupInfoTN
}

// CreateTable 创建表
func (info *GroupInfoDao_) CreateTable() {
	_ = info.DB.AutoMigrate(&GroupInfo{})
}

// ----------------------------------

func (util *GroupInfoUtil_) GetDefaultPost() string {
	return "该群暂无群公告~"
}
