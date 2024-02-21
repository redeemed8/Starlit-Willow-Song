package models

import (
	"gorm.io/gorm"
	"jcpd.cn/user/internal/constants"
	"jcpd.cn/user/internal/options"
	"regexp"
	"time"
)

var JoinApplyDao JoinApplyDao_
var JoinApplyUtil JoinApplyUtil_

type JoinApplyDao_ struct{ DB *gorm.DB }
type JoinApplyUtil_ struct{}

func NewJoinApplyDao() {
	JoinApplyDao = JoinApplyDao_{DB: options.C.DB}
}

type ApplyStatus string

const (
	Pending  ApplyStatus = "0"
	Accepted ApplyStatus = "1"
	Rejected ApplyStatus = "2"
	NotKnown ApplyStatus = "no"
)

func (type_ ApplyStatus) ToString() string {
	switch type_ {
	case Pending:
		return "等待验证"
	case Accepted:
		return "已添加"
	case Rejected:
		return "验证失败"
	}
	return string(type_)
}

const (
	Friend = "friend"
	Group  = "group"
)

// JoinApply 加入申请表，采用4个索引而不是联合索引，是因为这个表要经常写入和修改
// 而且数据量大时，联合索引占的内存可能比普通索引多，而且普通索引使用更灵活
type JoinApply struct {
	Id           uint32      `gorm:"primaryKey;autoIncrement"` //	请求 id作为主键
	CreatedAt    time.Time   //	创建时间
	UpdatedAt    time.Time   //	更新时间
	SenderId     uint32      `gorm:"not null;index:idx_field1"` //	发送人 id
	ReceiverId   uint32      `gorm:"not null;index:idx_field2"` //	接收者 id -- 人或群
	ApplyType    string      `gorm:"not null;index:idx_field3"` //	申请类型  添加好友或加入群聊
	Status       ApplyStatus `gorm:"not null;index:idx_field4"` //	申请状态
	Introduction string      `gorm:"not null"`                  //	申请人介绍
}

const JoinApplyTN = "5176_apply"

// TableName 表名
func (table *JoinApply) TableName() string {
	return JoinApplyTN
}

// CreateTable 创建表
func (join *JoinApplyDao_) CreateTable() {
	_ = join.DB.AutoMigrate(&JoinApply{})
}

// CreateApply 添加申请信息
func (join *JoinApplyDao_) CreateApply(apply JoinApply) error {
	return join.DB.Create(&apply).Error
}

// GetApplyByMap 根据 指定的信息获取 一条申请信息
func (join *JoinApplyDao_) GetApplyByMap(columnMap map[string]interface{}) (JoinApply, error) {
	var applies JoinApply
	result := join.DB.Model(&JoinApply{}).Where(columnMap).First(&applies)
	return applies, result.Error
}

// GetAppliesByMap 根据 指定的信息获取 多条申请信息
func (join *JoinApplyDao_) GetAppliesByMap(columnMap map[string]interface{}) ([]JoinApply, error) {
	var applies []JoinApply
	result := join.DB.Model(&JoinApply{}).Where(columnMap).Find(&applies)
	return applies, result.Error
}

func (join *JoinApplyDao_) UpdateApplyByMap(id uint32, columnMap map[string]interface{}) error {
	return join.DB.Model(&JoinApply{}).Where("id = ?", id).Updates(columnMap).Error
}

// ----------------------------------

const (
	//	此处应与客户端商量需要什么样的
	PPP = "0"
	AAA = "1"
	RRR = "2"
)

func (util *JoinApplyUtil_) TransToStatus(status string) ApplyStatus {
	switch status {
	case PPP:
		return Pending
	case AAA:
		return Accepted
	case RRR:
		return Rejected
	}
	return NotKnown
}

func (util *JoinApplyUtil_) GetDftIntroduce() string {
	return "Hi~ 我想成为你的朋友~,可以吗？"
}

func (util *JoinApplyUtil_) CheckIntroduce(introduce *string) bool {
	if *introduce == "" {
		*introduce = util.GetDftIntroduce()
		return true
	}
	if len(*introduce) > 300 {
		return false
	}
	return regexp.MustCompile(constants.IntroduceRegex).MatchString(*introduce)
}
