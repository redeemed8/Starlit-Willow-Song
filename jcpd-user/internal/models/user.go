package models

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	common "jcpd.cn/common/models"
	commonJWT "jcpd.cn/common/utils/jwt"
	"jcpd.cn/user/internal/constants"
	"jcpd.cn/user/internal/models/dto"
	"jcpd.cn/user/internal/options"
	"jcpd.cn/user/pkg/definition"
	"jcpd.cn/user/utils"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var UserInfoDao UserInfoDao_
var UserInfoUtil UserInfoUtil_

type UserInfoDao_ struct{ DB *gorm.DB }
type UserInfoUtil_ struct{}

func NewUserInfoDao() {
	UserInfoDao = UserInfoDao_{DB: options.C.DB}
}

const (
	DefaultSex = "2"
	Man        = "1"
	Woman      = "0"
)

type UserInfo struct {
	Id         uint32    `gorm:"primaryKey;autoIncrement"` //	主键 id
	Phone      string    `gorm:"size:12;unique"`           //	手机号 - 唯一
	Username   string    `gorm:"size:31;unique"`           //	用户名 - 唯一
	Password   string    `gorm:"size:33"`                  //	密码 - md5存储
	UUID       string    `gorm:"size:37;not null"`         //	用户身份标识 - 存储在jwt中，会随着密码的修改而修改
	Sex        string    `gorm:"size:2"`                   //	性别   0女  1男  2未知
	Sign       string    `gorm:"type:longtext"`            //	个性签名
	FriendList string    `gorm:"type:longtext"`            //	好友列表
	GroupList  string    `gorm:"type:longtext"`            //	群聊列表
	CreatedAt  time.Time `gorm:"autoCreateTime"`           //	创建时间
}

const UserInfoTN = "5613_userinfo"

// TableName 表名
func (table *UserInfo) TableName() string {
	return UserInfoTN
}

// CreateTable 创建表
func (info *UserInfoDao_) CreateTable() {
	_ = info.DB.AutoMigrate(&UserInfo{})
}

// CreateUser 创建用户信息
func (info *UserInfoDao_) CreateUser(userinfo UserInfo) error {
	return info.DB.Create(&userinfo).Error
}

// GetUserById 根据用户id获取用户信息
func (info *UserInfoDao_) GetUserById(id uint32) (UserInfo, error) {
	userinfo := UserInfo{}
	result := info.DB.Where("id = ?", id).First(&userinfo)
	return userinfo, result.Error
}

// GetUsersByIds 根据 id获取部分用户的部分信息
func (info *UserInfoDao_) GetUsersByIds(ids []uint32) ([]UserInfo, error) {
	idsStr := UserInfoUtil.JoinUint32(ids)
	//	select username,sex,sign from UserInfoTN where id in (...)
	sqlSlice := []string{
		"select id,username,sex,sign from",
		UserInfoTN,
		fmt.Sprintf("where id in (%s)", idsStr),
	}
	sql_ := strings.Join(sqlSlice, " ")
	var infos []UserInfo
	err := info.DB.Raw(sql_).Scan(&infos).Error
	return infos, err
}

// GetUserByUsername 根据 用户名 获取用户信息
func (info *UserInfoDao_) GetUserByUsername(username string) (UserInfo, error) {
	userinfo := UserInfo{}
	result := info.DB.Where("username = ?", username).First(&userinfo)
	return userinfo, result.Error
}

// GetUserByPhone 根据 手机号 获取用户信息
func (info *UserInfoDao_) GetUserByPhone(phone string) (UserInfo, error) {
	userinfo := UserInfo{}
	result := info.DB.Where("phone = ?", phone).First(&userinfo)
	return userinfo, result.Error
}

// GetUsersByMap 根据 指定字段值 获取 一个或多个用户信息
func (info *UserInfoDao_) GetUsersByMap(columnMap map[string]interface{}) ([]UserInfo, error) {
	var userinfos []UserInfo
	result := info.DB.Where(columnMap).Find(userinfos)
	return userinfos, result.Error
}

// UpdateUserByMap 根据 id 更新map里的指定列
func (info *UserInfoDao_) UpdateUserByMap(id uint32, columnMap map[string]interface{}) error {
	return info.DB.Model(&UserInfo{}).Where("id = ?", id).Updates(columnMap).Error
}

// UpdateUser 根据 拥有UserInfo的部分字段的结构体来更新字段
func (info *UserInfoDao_) UpdateUser(id uint32, anyInfo interface{}) error {
	return info.DB.Model(&UserInfo{}).Where("id = ?", id).Updates(anyInfo).Error
}

// ----------------------------------

// CheckUsername 检查用户名是否合法
func (util *UserInfoUtil_) CheckUsername(username string) bool {
	if len(username) > 30 || username == "" {
		return false
	}
	return regexp.MustCompile(constants.UsernameRegex).MatchString(username)
}

// CheckSign 检查个性签名是否合法
func (util *UserInfoUtil_) CheckSign(sign string) bool {
	if len(sign) > 150 || sign == "" {
		return false
	}
	return regexp.MustCompile(constants.SignRegex).MatchString(sign)
}

// GetDefaultSex 获取默认性别
func (util *UserInfoUtil_) GetDefaultSex() string {
	return DefaultSex
}

const DefaultNamePrefix = "LXY"

// GetDefaultName 获取默认用户名
func (util *UserInfoUtil_) GetDefaultName() string {
	return DefaultNamePrefix + utils.MakeCodeWithNumber(11, rand.Intn(10))
}

// IsLogin 是否登录
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

// TransSex 性别转换
func (util *UserInfoUtil_) TransSex(sexCode string) string {
	if sexCode == Man {
		return "男"
	} else if sexCode == Woman {
		return "女"
	}
	return "未知"
}

func (util *UserInfoUtil_) transToDto(userinfo UserInfo) dto.UserInfoDto {
	var dto_ dto.UserInfoDto
	err := copier.Copy(&dto_, &userinfo)
	if err != nil {
		log.Printf("Failed to copy struct , source == %v , dest == %v , err == %v ... \n", userinfo, dto_, err)
	}
	return dto_
}

func (util *UserInfoUtil_) TransToDtos(userinfos ...UserInfo) dto.UserInfoDtos {
	var dtos dto.UserInfoDtos
	for _, info := range userinfos {
		dtos = append(dtos, util.transToDto(info))
	}
	return dtos
}

func (util *UserInfoUtil_) JoinUint32(ids []uint32) string {
	var idsStr string
	for i, id := range ids {
		idsStr = fmt.Sprintf("%s%d", idsStr, id)
		if i+1 == len(ids) {
			break
		}
		idsStr += ","
	}
	return idsStr
}

func (util *UserInfoUtil_) IdIsExists(ids string, id uint32) bool {
	idStr := fmt.Sprintf("%d", id)
	idArr := strings.Split(ids, ",")
	for _, tId := range idArr {
		if tId == idStr {
			return true
		}
	}
	return false
}

func (util *UserInfoUtil_) AddToList(list *string, target string) {
	if *list == "" {
		*list += target
	} else if target != "" {
		*list += "," + target
	}
}
