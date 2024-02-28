package models

import (
	"gorm.io/gorm"
	"jcpd.cn/user/internal/constants"
	"jcpd.cn/user/internal/models/dto"
	"jcpd.cn/user/internal/options"
	"jcpd.cn/user/utils"
	"regexp"
	"strconv"
	"strings"
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
	AdminIds     string    `gorm:"type:text"`                //	管理员 id
	MemberIds    string    `gorm:"type:text"`                //	成员 id
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

// CreateGroup 创建一个群信息
func (info *GroupInfoDao_) CreateGroup(groupInfo *GroupInfo) error {
	return info.DB.Create(groupInfo).Error
}

// GetGroupInfoById 根据 id获取群信息
func (info *GroupInfoDao_) GetGroupInfoById(id uint32) (GroupInfo, error) {
	var group GroupInfo
	result := info.DB.Model(&GroupInfo{}).Where("id = ?", id).First(&group)
	return group, result.Error
}

// GetGroupsByIds 根据id批量查询群信息
func (info *GroupInfoDao_) GetGroupsByIds(ids []uint32) (GroupInfos, error) {
	var infos GroupInfos
	result := info.DB.Model(&GroupInfo{}).Where("id in ?", ids).Find(&infos)
	return infos, result.Error
}

// GetGroupInfoByName 根据 name获取群信息
func (info *GroupInfoDao_) GetGroupInfoByName(groupName string) (GroupInfo, error) {
	var group GroupInfo
	result := info.DB.Model(&GroupInfo{}).Where("group_name = ?", groupName).First(&group)
	return group, result.Error
}

// UpdateGroup 根据 GroupInfo更新群信息
func (info *GroupInfoDao_) UpdateGroup(id uint32, groupInfo GroupInfo) error {
	return info.DB.Model(&GroupInfo{}).Where("id = ?", id).Updates(groupInfo).Error
}

// UpdateGroupByMap 根据 Map更新群信息
func (info *GroupInfoDao_) UpdateGroupByMap(id uint32, columnMap map[string]interface{}) error {
	return info.DB.Model(&GroupInfo{}).Where("id = ?", id).Updates(columnMap).Error
}

// DeleteGroupById 根据id删除群信息
func (info *GroupInfoDao_) DeleteGroupById(id uint32) error {
	return info.DB.Model(&GroupInfo{}).Where("id = ?", id).Delete(&GroupInfo{}).Error
}

// ----------------------------------

type GroupInfos []GroupInfo

func (groups *GroupInfos) Names() []string {
	var names []string
	for _, group := range *groups {
		names = append(names, group.GroupName)
	}
	return names
}

// GetDefaultPost 获取默认的群公告
func (util *GroupInfoUtil_) GetDefaultPost() string {
	return "该群暂无群公告~"
}

// CheckGroupName 检查群名称
func (util *GroupInfoUtil_) CheckGroupName(name *string) bool {
	if *name == "" || len(*name) > 45 {
		return false
	}
	return regexp.MustCompile(constants.GroupNameRegex).MatchString(*name)
}

// CheckGroupPost 检查群公告
func (util *GroupInfoUtil_) CheckGroupPost(post *string) bool {
	if *post == "" {
		*post = util.GetDefaultPost()
		return true
	}
	if len(*post) > 1000 {
		return false
	}
	return regexp.MustCompile(constants.GroupPostRegex).MatchString(*post)
}

// CheckGroupMaxNum 检查群人数
func (util *GroupInfoUtil_) CheckGroupMaxNum(maxNum *int) bool {
	if *maxNum > 500 {
		*maxNum = 500
		return true
	}
	return *maxNum > 0
}

// parseIds 将id字符串转换为数组
func (util *GroupInfoUtil_) parseIds(idStr_ *string) []uint32 {
	idStrArr := strings.Split(*idStr_, ",")
	if *idStr_ == "" || len(idStrArr) == 0 {
		return make([]uint32, 0)
	}
	ids := make([]uint32, 0) //	结果
	for _, idStr := range idStrArr {
		//	转换为 uint32
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			continue
		}
		ids = append(ids, uint32(id))
	}
	return ids
}

// TransToDto 封装为 dto
func (util *GroupInfoUtil_) TransToDto(group GroupInfo) dto.GroupInfoDto {
	return dto.GroupInfoDto{
		Id:           group.Id,
		GroupName:    group.GroupName,
		GroupPost:    group.GroupPost,
		LordId:       group.LordId,
		AdminIds:     util.parseIds(&(group.AdminIds)),
		MemberIds:    util.parseIds(&(group.MemberIds)),
		CurPersonNum: group.CurPersonNum,
		MaxPersonNum: group.MaxPersonNum,
	}
}

// IsAdmin 检查一个用户是否是某群的管理员
func (util *GroupInfoUtil_) IsAdmin(groupInfo GroupInfo, userId uint32) bool {
	idStr := strconv.Itoa(int(userId))
	ids := strings.Split(groupInfo.AdminIds, ",")
	for _, id := range ids {
		if id == idStr {
			return true
		}
	}
	return false
}

// DeleteFromList 从id列表中删除某个id
func (util *GroupInfoUtil_) DeleteFromList(list *string, targetId uint32) {
	UintIds := utils.ParseListToUint(*list)
	for i := range UintIds {
		if UintIds[i] == targetId {
			utils.RemoveIdFromList(&UintIds, i)
		}
	}
	*list = utils.JoinUint32(UintIds)
}
