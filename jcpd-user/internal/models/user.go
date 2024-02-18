package models

import (
	"jcpd.cn/user/internal/options"
	"time"
)

var UserInfoDao UserInfoDao_

type UserInfoDao_ struct {
}

type UserInfo struct {
	Id        uint32    `gorm:"primaryKey;autoIncrement"` //	主键 id
	Phone     string    `gorm:"unique"`                   //	手机号 - 唯一
	Username  string    `gorm:"unique"`                   //	用户名 - 唯一
	Password  string    //	密码 - md5存储
	UUID      string    `gorm:"not null"` //	用户身份标识 - 存储在jwt中，会随着密码的修改而修改
	Sex       string    //	性别   0女  1男  2未知
	Sign      string    //	个性签名
	CreatedAt time.Time //	创建时间
}

func (table *UserInfo) TableName() string {
	return "5613_userinfo"
}

func (info *UserInfoDao_) CreateUser(userinfo UserInfo) error {
	return options.C.DB.Create(&userinfo).Error
}
