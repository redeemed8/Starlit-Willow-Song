package models

import (
	"errors"
	"github.com/gin-gonic/gin"
	common "jcpd.cn/common/models"
	commonJWT "jcpd.cn/common/utils/jwt"
	"jcpd.cn/user/internal/constants"
	"jcpd.cn/user/internal/options"
	"jcpd.cn/user/pkg/definition"
	"jcpd.cn/user/utils"
	"math/rand"
	"net/http"
	"regexp"
	"time"
)

var UserInfoDao UserInfoDao_
var UserInfoUtil UserInfoUtil_

type UserInfoDao_ struct{}
type UserInfoUtil_ struct{}

const (
	DefaultSex = "2"
	Man        = "1"
	Woman      = "0"
)

type UserInfo struct {
	Id        uint32    `gorm:"primaryKey;autoIncrement"` //	主键 id
	Phone     string    `gorm:"size:12;unique"`           //	手机号 - 唯一
	Username  string    `gorm:"size:31;unique"`           //	用户名 - 唯一
	Password  string    `gorm:"size:33"`                  //	密码 - md5存储
	UUID      string    `gorm:"size:37;not null"`         //	用户身份标识 - 存储在jwt中，会随着密码的修改而修改
	Sex       string    `gorm:"size:2"`                   //	性别   0女  1男  2未知
	Sign      string    `gorm:"type:longtext;"`           //	个性签名
	CreatedAt time.Time `gorm:"autoCreateTime"`           //	创建时间
}

// TableName 表名
func (table *UserInfo) TableName() string {
	return "5613_userinfo"
}

// CreateTable 创建表
func (info *UserInfoDao_) CreateTable() {
	_ = options.C.DB.AutoMigrate(&UserInfo{})
}

// CreateUser 创建用户信息
func (info *UserInfoDao_) CreateUser(userinfo UserInfo) error {
	return options.C.DB.Create(&userinfo).Error
}

// GetUserById 根据用户id获取用户信息
func (info *UserInfoDao_) GetUserById(id uint32) (UserInfo, error) {
	userinfo := UserInfo{}
	result := options.C.DB.Where("id = ?", id).First(&userinfo)
	return userinfo, result.Error
}

// GetUserByUsername 根据 用户名 获取用户信息
func (info *UserInfoDao_) GetUserByUsername(username string) (UserInfo, error) {
	userinfo := UserInfo{}
	result := options.C.DB.Where("username = ?", username).First(&userinfo)
	return userinfo, result.Error
}

// GetUserByPhone 根据 手机号 获取用户信息
func (info *UserInfoDao_) GetUserByPhone(phone string) (UserInfo, error) {
	userinfo := UserInfo{}
	result := options.C.DB.Where("phone = ?", phone).First(&userinfo)
	return userinfo, result.Error
}

// GetUsersByMap 根据 指定字段值 获取 一个或多个用户信息
func (info *UserInfoDao_) GetUsersByMap(columnMap map[string]interface{}) ([]UserInfo, error) {
	var userinfos []UserInfo
	result := options.C.DB.Where(columnMap).Find(userinfos)
	return userinfos, result.Error
}

// UpdateUserByMap 根据 id 更新map里的指定列
func (info *UserInfoDao_) UpdateUserByMap(id uint32, columnMap map[string]interface{}) error {
	return options.C.DB.Model(&UserInfo{}).Where("id = ?", id).Updates(columnMap).Error
}

// ----------------------------------

// CheckUsername 检查用户名是否合法
func (util *UserInfoUtil_) CheckUsername(username string) bool {
	if len(username) > 30 || username == "" {
		return false
	}
	return regexp.MustCompile(constants.UsernameRegex).MatchString(username)
}

func (util *UserInfoUtil_) GetDefaultSex() string {
	return DefaultSex
}

func (util *UserInfoUtil_) GetDefaultName() string {
	return "用户" + utils.MakeCodeWithNumber(11, rand.Intn(10))
}

func (util *UserInfoUtil_) IsLogin(ctx *gin.Context, resp *common.Resp) (*common.NormalErr, commonJWT.UserClaims) {
	userClaims, err := commonJWT.ParseToken(ctx)
	if errors.Is(err, commonJWT.DBException) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.ServerError))
		return &definition.ServerError, userClaims
	}
	if errors.Is(err, commonJWT.NotLoginError) {
		ctx.JSON(http.StatusOK, resp.Fail(definition.NotLogin))
		return &definition.NotLogin, userClaims
	}
	return nil, userClaims
}

func (util *UserInfoUtil_) TransSex(sexCode string) string {
	if sexCode == Man {
		return "男"
	} else if sexCode == Woman {
		return "女"
	}
	return "未知"
}
