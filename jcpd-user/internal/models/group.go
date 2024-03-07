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
	BlackList    string    `gorm:"type:text"`                //	黑名单
	CreatedAt    time.Time //	创建时间
	UpdatedAt    time.Time //	更新时间
	Status       string    `gorm:"default:'ok';index"` // 群信息状态 -- ok表示正常，deleted表示群已被解散
}

const GroupInfoTN = "5433_group"
const GroupDeleted = "deleted"
const GroupOK = "ok"

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

// GetGroupsByMap 根据指定条件获取
func (info *GroupInfoDao_) GetGroupsByMap(condition map[string]interface{}) (GroupInfos, error) {
	infos := make(GroupInfos, 0)
	result := info.DB.Model(&GroupInfo{}).Where(condition).Find(&infos)
	return infos, result.Error
}

// GetGroupsByIds 根据id批量查询群信息
func (info *GroupInfoDao_) GetGroupsByIds(ids []uint32) (GroupInfos, error) {
	var infos GroupInfos
	result := info.DB.Model(&GroupInfo{}).Where("id in ?", ids).Find(&infos)
	return infos, result.Error
}

// GetGroupInfoByName 根据 name获取群信息
func (info *GroupInfoDao_) GetGroupInfoByName(groupName string) (*GroupInfos, error) {
	groups := make(GroupInfos, 0)
	result := info.DB.Model(&GroupInfo{}).Where("group_name = ?", groupName).First(&groups)
	return &groups, result.Error
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
func (info *GroupInfoDao_) DeleteGroupById(ids []uint32) error {
	return info.DB.Model(&GroupInfo{}).Where("id in ?", ids).Delete(&GroupInfo{}).Error
}

// DeleteGroupByMap 根据 指定字段 删除群信息
func (info *GroupInfoDao_) DeleteGroupByMap(condition map[string]interface{}) error {
	return info.DB.Model(&GroupInfo{}).Where(condition).Delete(&GroupInfo{}).Error
}

// ----------------------------------

type GroupInfos []GroupInfo

// Names 获取所有的群名称
func (groups *GroupInfos) Names() []string {
	var names []string
	for _, group := range *groups {
		if group.GroupName != "" {
			names = append(names, group.GroupName)
		}
	}
	return names
}

func (groups *GroupInfos) Ids() []uint32 {
	var ids []uint32
	for _, group := range *groups {
		if group.GroupName != "" {
			ids = append(ids, group.Id)
		}
	}
	return ids
}

// RemoveDeleted 排除所有已被删除的
func (groups *GroupInfos) RemoveDeleted() *GroupInfos {
	groupArr := make(GroupInfos, 0)
	for _, group := range *groups {
		if group.Status == GroupOK && group.GroupName != "" {
			groupArr = append(groupArr, group)
		}
	}
	return &groupArr
}

// GetDefaultPost 获取默认的群公告
func (util *GroupInfoUtil_) GetDefaultPost() string {
	return "该群暂无群公告~"
}

// GetDeletedStatus 获取群被删除的状态
func (util *GroupInfoUtil_) GetDeletedStatus() string {
	return GroupDeleted
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
		if idStr == "" {
			continue
		}
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

// TransToDtos 批量封装为 dto
func (util *GroupInfoUtil_) TransToDtos(groups GroupInfos) []dto.GroupInfoDto {
	var ret []dto.GroupInfoDto
	for _, group := range groups {
		ret = append(ret, util.TransToDto(group))
	}
	return ret
}

// IsAdmin 检查一个用户是否是某群的管理员
func (util *GroupInfoUtil_) IsAdmin(groupInfo GroupInfo, userId uint32) bool {
	return utils.FindIdFromIdsStr(groupInfo.AdminIds, userId)
}

// IsMember 检查一个用户是否是某群的成员
func (util *GroupInfoUtil_) IsMember(groupInfo GroupInfo, userId uint32) bool {
	return utils.FindIdFromIdsStr(groupInfo.MemberIds, userId)
}

// IsExistBlackList 检查一个用户是否存在于群的黑名单中
func (util *GroupInfoUtil_) IsExistBlackList(groupInfo GroupInfo, userId uint32) bool {
	return utils.FindIdFromIdsStr(groupInfo.BlackList, userId)
}

// AddToList 添加某个 id串到 id列表中
func (util *GroupInfoUtil_) AddToList(list *string, target string) string {
	*list += target + ","
	return *list
}

// DeleteFromList 从id列表中删除某个id
func (util *GroupInfoUtil_) DeleteFromList(list *string, targetId uint32) string {
	target := "," + strconv.Itoa(int(targetId)) + ","
	*list = strings.Replace(*list, target, ",", 1)
	return *list
}
