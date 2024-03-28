package models

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"jcpd.cn/talk/internal/constants"
	"jcpd.cn/talk/internal/options"
	"log"
)

var MessageCounterDao MessageCounterDao_
var MessageCounterUtil MessageCounterUtil_

type MessageCounterDao_ struct{ DB *gorm.DB }
type MessageCounterUtil_ struct{}

func NewMessageCounterDao() {
	MessageCounterDao = MessageCounterDao_{DB: options.C.DB}
}

type MessageCounter struct {
	Id         int    `gorm:"primaryKey;autoIncrement"`  //  主键 id
	SendToRece string `gorm:"size:22;index:send_rece"`   //	发送人和接收人的id组合
	UnreadNum  uint16 `gorm:"default:0;index:send_rece"` //	未读消息数
}

const MessageCounterTN = "1455_message_counter"

// TableName 表名
func (counter *MessageCounter) TableName() string {
	return MessageCounterTN
}

// CreateTable 创建表
func (dao *MessageCounterDao_) CreateTable() {
	_ = dao.DB.AutoMigrate(&MessageCounter{})
}

// CreateMessageCounter	 创建记录
func (dao *MessageCounterDao_) CreateMessageCounter(senderId uint32, receiverId uint32) error {
	if _, err := dao.GetUnreadCounter(senderId, receiverId); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return dao.DB.Model(&MessageCounter{}).
		Create(&MessageCounter{SendToRece: MessageCounterUtil.MakeSendToRece(senderId, receiverId), UnreadNum: 0}).Error
}

// GetUnreadCounter  获取未读消息数
func (dao *MessageCounterDao_) GetUnreadCounter(senderId uint32, receiverId uint32) (uint16, error) {
	var unreadNum uint16
	result := dao.DB.Model(&MessageCounter{}).Select("unread_num").
		Where("send_to_rece = ?", MessageCounterUtil.MakeSendToRece(senderId, receiverId)).First(&unreadNum)
	return unreadNum, result.Error
}

// UpdateMessageCounter  修改未读消息数
func (dao *MessageCounterDao_) UpdateMessageCounter(senderId uint32, receiverId uint32, toCount uint16) error {
	return dao.DB.Model(&MessageCounter{}).
		Where("send_to_rece = ?", MessageCounterUtil.MakeSendToRece(senderId, receiverId)).
		Update("unread_num", toCount).Error
}

// AddMessageCounter  将未读消息数+1
func (dao *MessageCounterDao_) AddMessageCounter(senderId uint32, receiverId uint32) error {
	return dao.DB.Model(&MessageCounter{}).
		Where("send_to_rece = ?", MessageCounterUtil.MakeSendToRece(senderId, receiverId)).
		UpdateColumn("unread_num", gorm.Expr("unread_num + ?", 1)).Error
}

// CountToZero  将未读消息数置零
func (dao *MessageCounterDao_) CountToZero(senderId uint32, receiverId uint32) error {
	//	清零之前要将message表中的status改为已读
	condition := map[string]interface{}{"sender_id": senderId, "receiver_id": receiverId}
	updates := map[string]interface{}{"status": Readed}
	err := MessageInfoDao.UpdateMessage(condition, updates)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println(constants.Err("修改消息阅读状态出错 , cause by :" + err.Error()))
		return errors.New("1")
	}
	return dao.UpdateMessageCounter(senderId, receiverId, 0)
}

// ------------------------------------------------------

const SendToReceSplit = "-"

func (util *MessageCounterUtil_) MakeSendToRece(senderId uint32, receiverId uint32) string {
	return fmt.Sprintf("%d%s%d", senderId, SendToReceSplit, receiverId)
}
