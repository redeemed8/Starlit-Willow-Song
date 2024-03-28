package models

import (
	"errors"
	"gorm.io/gorm"
	"jcpd.cn/talk/internal/constants"
	"jcpd.cn/talk/internal/options"
	"log"
	"time"
)

var MessageInfoDao MessageInfoDao_
var MessageInfoUtil MessageInfoUtil_

type MessageInfoDao_ struct{ DB *gorm.DB }
type MessageInfoUtil_ struct{}

func NewMessageInfoDao() {
	MessageInfoDao = MessageInfoDao_{DB: options.C.DB}
}

type Message struct {
	Id         int       `gorm:"primaryKey;autoIncrement"` //  主键 id
	CreatedAt  time.Time `gorm:"autoCreateTime"`           //  创建时间
	SenderId   uint32    `gorm:"index:message_id"`         //  消息发送人id
	ReceiverId uint32    `gorm:"index:message_id"`         //  消息接收人id
	Content    string    `gorm:"type:text"`                //  消息体
	Status     string    `gorm:"size:1"`                   //  阅读状态
}

const MessageInfoTN = "1765_message"
const Unread = "0"
const Readed = "1"
const ForwardSort = ""
const ReverseSort = " DESC"

// TableName 表名
func (message *Message) TableName() string {
	return MessageInfoTN
}

type Messages []Message

func (msgs *Messages) ToStrings() []string {
	var contents = make([]string, 0)
	for _, msg := range *msgs {
		contents = append(contents, msg.Content)
	}
	return contents
}

// CreateTable 创建表
func (info *MessageInfoDao_) CreateTable() {
	_ = info.DB.AutoMigrate(&Message{})
}

var UnknownMessageStatus = errors.New("UnknownMessageStatus")

const ALL = -10086

// GetMessage  根据对话双方id，获取消息内容，默认为正向排序，默认获取50条
func (info *MessageInfoDao_) GetMessage(curUserId uint32, targetId uint32, sortBy string, status string, limit int) ([]string, error) {
	var msgs = make(Messages, 0)
	//	排序方式
	if sortBy == ReverseSort {
		sortBy = "created_at" + ReverseSort
	} else {
		sortBy = "created_at" + ForwardSort
	}
	//	阅读状态
	if status != Unread && status != Readed {
		return make([]string, 0), UnknownMessageStatus
	}
	//	获取数量
	if limit < 1 && limit != ALL {
		limit = 50
	}
	//	查询数据库
	tx := info.DB.Model(&Message{}).Where("sender_id = ? AND receiver_id = ? AND status = ?", curUserId, targetId, status)
	if limit != ALL {
		tx = tx.Limit(limit)
	}
	result := tx.Order(sortBy).Find(&msgs)
	return msgs.ToStrings(), result.Error
}

// CreateMessage  创建一条新消息，如果是未读消息应该将未读数+1
func (info *MessageInfoDao_) CreateMessage(message *Message) error {
	err := info.DB.Model(&Message{}).Create(message).Error
	if err == nil && message.Status == Unread {
		//	因为记录已经创建，所以直接修改未读数+1即可
		err0 := MessageCounterDao.AddMessageCounter(message.SenderId, message.ReceiverId)
		if err0 != nil && !errors.Is(err0, gorm.ErrRecordNotFound) {
			log.Println(constants.Err("递增消息未读数出错 , cause by : " + err0.Error()))
		}
	}
	return err
}

// UpdateMessage  修改message信息
func (info *MessageInfoDao_) UpdateMessage(condition map[string]interface{}, updates map[string]interface{}) error {
	return info.DB.Model(&Message{}).Where(condition).Updates(updates).Error
}

// ------------------------------------------------------
