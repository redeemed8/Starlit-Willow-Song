package models

import (
	"errors"
	"gorm.io/gorm"
	common "jcpd.cn/common/models"
	"jcpd.cn/post/internal/models/dto"
	"jcpd.cn/post/internal/options"
	"jcpd.cn/post/pkg/definition"
	"strconv"
	"time"
)

var PostInfoDao postInfoDao_
var PostInfoUtil postInfoUtil_

type postInfoDao_ struct{ DB *gorm.DB }
type postInfoUtil_ struct{}

func NewPostInfoDao() {
	PostInfoDao = postInfoDao_{DB: options.C.DB}
}

// PostInfo 帖子类
type PostInfo struct {
	Id            uint32    `gorm:"primaryKey;autoIncrement"` //	主键 id -- 帖子id
	CreatedAt     time.Time //	帖子创建时间
	UpdatedAt     time.Time `gorm:"index:like_time"`            //	帖子的最近一次修改时间
	Title         string    `gorm:"not null;type:text"`         //	帖子标题
	TopicTag      string    `gorm:"not null;size:60;index:ttt"` //	主题标签
	Body          string    `gorm:"not null;type:text"`         //	帖子内容
	PublisherId   uint32    `gorm:"not null;index:ppp"`         //	发布人id
	PublisherName string    `gorm:"not null;size:31"`           //	发布人用户名
	Likes         int       `gorm:"default:0;index:like_time"`  //	点赞数 - 热度
	Comments      int       `gorm:"default:0"`                  //	评论数
	Favorites     int       `gorm:"default:0"`                  //	收藏数
	ReviewStatus  string    `gorm:"size:1;default:'0'"`         //	审核状态, 0-未审核，1-已通过，2-已驳回
	Reason        string    //	驳回原因 -- 保存3天
}

const PostInfoTN = "3491_postinfo"

// TableName 表名
func (post *PostInfo) TableName() string {
	return PostInfoTN
}

// CreateTable 创建表
func (info *postInfoDao_) CreateTable() {
	_ = info.DB.AutoMigrate(&PostInfo{})
}

// CreatePost 创建帖子信息
func (info *postInfoDao_) CreatePost(post *PostInfo) error {
	return info.DB.Model(&PostInfo{}).Create(post).Error
}

// SimpleGetPostsPage 简单的分页查询，仅用了联合索引对排序列进行了优化
func (info *postInfoDao_) SimpleGetPostsPage(pageargs PageArgs) (PostInfos, error) {
	//	分页 - 热度优先，时间次之
	infos := make(PostInfos, 0)
	result := info.DB.Model(&PostInfo{}).
		Where("created_at >= ?", time.Now().AddDate(0, -1, 0)). //	 优先获取最近一个月内的,不能说一个视频热就一直热
		Order("likes DESC,created_at DESC").
		Limit(pageargs.PageSize).
		Offset((pageargs.PageNum - 1) * pageargs.PageSize).
		Find(&infos)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return infos, result.Error
	}
	if infos.size() == 0 { //	此处说明跳过的帖子太多，将一个月内的都跳过了，我们就查一个月前的
		result = info.DB.Model(&PostInfo{}).
			Order("likes DESC,created_at DESC").
			Limit(pageargs.PageSize).
			Offset((pageargs.PageNum - 1) * pageargs.PageSize).
			Find(&infos)
	}
	return infos, result.Error
}

// SeniorGetPostPage 优化后的分页查询，将 使用offset偏移量的方式，更改为用 where条件过滤的方式
// 因为后发的帖子id大，所以可以根据id来进行过滤
func (info *postInfoDao_) SeniorGetPostPage(pageargs PageArgs, lastMinPostId uint32, ok bool) (PostInfos, error) {
	//	分页 - 时间优先，热度次之
	infos := make(PostInfos, 0)
	tx := info.DB.Model(&PostInfo{}) //	 不变 sql
	if ok {
		//  说明不是第一次查询，我们为其优化
		tx = tx.Where("id < ?", lastMinPostId)
	}
	result := tx.Order("created_at DESC,likes DESC").Limit(pageargs.PageSize).Find(&infos) //	 不变 sql
	return infos, result.Error
}

// --------------------------------------------------

type ReviewStatus string

var (
	Wait = [2]string{"0", "等待审核"}
	OK   = [2]string{"1", "审核已通过"}
	Fail = [2]string{"2", "帖子被驳回"}
)

func (status *ReviewStatus) ToString() string {
	switch *status {
	case ReviewStatus(Wait[0]):
		return Wait[1]
	case ReviewStatus(OK[0]):
		return OK[1]
	case ReviewStatus(Fail[0]):
		return Fail[1]
	}
	return Wait[1]
}

// --------------------------------------------------

type PostInfos []PostInfo

func (infos *PostInfos) size() int {
	return len(*infos)
}

func (infos *PostInfos) ToDtos() []dto.PostInfoDto {
	var dtos = make([]dto.PostInfoDto, len(*infos))
	for i, info := range *infos {
		dtos[i] = PostInfoUtil.TransToDto(info)
	}
	return dtos
}

// --------------------------------------------------

const (
	TitleWordCount = 50 //	这里均以汉字计数
	TopicWordCount = 20
	BodyWordCount  = 1500
)

// CheckPostTitle 检查帖子标题
func (util *postInfoUtil_) CheckPostTitle(title string) bool {
	if title == "" || len(title) > TitleWordCount*3 {
		return false
	}
	return true
}

// CheckPostTopicTag 检查帖子主题标签
func (util *postInfoUtil_) CheckPostTopicTag(topicTag string) bool {
	if topicTag == "" || len(topicTag) > TopicWordCount*3 {
		return false
	}
	return true
}

// CheckPostBody 检查帖子内容
func (util *postInfoUtil_) CheckPostBody(body string) bool {
	if body == "" || len(body) > BodyWordCount*3 {
		return false
	}
	return true
}

// CheckPostBase 检查帖子主体内容
func (util *postInfoUtil_) CheckPostBase(post PostInfo) *common.NormalErr {
	if ok := util.CheckPostTitle(post.Title); !ok {
		return &definition.PostTitleNotFormat
	}
	if ok := util.CheckPostTopicTag(post.TopicTag); !ok {
		return &definition.PostTopicNotFormat
	}
	if ok := util.CheckPostBody(post.Body); !ok {
		return &definition.PostBodyNotFormat
	}
	return nil
}

func (util *postInfoUtil_) CheckPage(pagenum string, pagesize string) (page PageArgs, retErr *common.NormalErr) {
	var err error
	var pageNum, pageSize int

	pageNum, err = strconv.Atoi(pagenum)
	if pagenum == "" || err != nil || pageNum < 1 {
		return PageArgs{-1, -1}, &definition.PageNumNotFormat
	}
	pageSize, err = strconv.Atoi(pagesize)
	if pagesize == "" || err != nil || pageSize < 0 {
		return PageArgs{-1, -1}, &definition.PageSizeNotFormat
	}
	return PageArgs{pageNum, pageSize}, nil
}

func (util *postInfoUtil_) TransToDto(info PostInfo) dto.PostInfoDto {
	return dto.PostInfoDto{
		Id:            info.Id,
		Title:         info.Title,
		TopicTag:      info.TopicTag,
		Body:          info.Body,
		PublisherName: info.PublisherName,
		PublishTime:   info.CreatedAt,
		Likes:         info.Likes,
		Comments:      info.Comments,
		Favorites:     info.Favorites,
		ReviewStatus:  info.ReviewStatus,
		Reason:        info.Reason,
	}
}

// CheckLmid 检查 lmid参数是否可以开启优化， 返回值-整型lmid，是否开启  --  默认为不开启
func (util *postInfoUtil_) CheckLmid(lmid string) (uint32, bool) {
	id, err := strconv.Atoi(lmid)
	if lmid == "" || err != nil {
		return 0, false
	}
	//	进行类型转换
	lastMinPostId := uint32(id)
	if int(lastMinPostId) != id {
		return 0, false
	}
	return lastMinPostId, true
}
