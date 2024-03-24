package service

import (
	"context"
	"errors"
	"gorm.io/gorm"
	common "jcpd.cn/common/models"
	"jcpd.cn/user/internal/models"
	"jcpd.cn/user/pkg/definition"
	"strconv"
	"strings"
)

type UserService struct {
	UnimplementedUserServiceServer
}

func New() *UserService {
	return &UserService{}
}

func (u *UserService) GetUserById(ctx context.Context, request *UserRequest) (*UserResponse, error) {
	queryUser, err := models.UserInfoDao.GetUserById(request.UserId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		//	服务降级...
		return nil, definition.ServerMaintaining.Err()
	}
	if queryUser.Username == "" {
		return nil, definition.UserNotFound.Err()
	}
	return &UserResponse{Username: queryUser.Username, Uuid: queryUser.UUID}, nil
}

func (u *UserService) IsRelated(ctx context.Context, request *UserRelationDecideRequest) (*UserRelationDecideResponse, error) {
	if request.TargetId < 1 {
		return nil, definition.UserNotFound.Err()
	}
	//	1. 查数据库列表
	queryUser, err := models.UserInfoDao.GetUserById(request.UserId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		//	服务降级...
		return nil, definition.ServerMaintaining.Err()
	}
	if queryUser.Username == "" {
		return nil, definition.UserNotFound.Err()
	}
	//	2. 对象id
	var target = strconv.Itoa(int(request.TargetId))
	if target == "" {
		return nil, definition.UserNotFound.Err()
	}
	//	3. 确定列表
	var stringList string
	if request.FORg == models.Friend {
		stringList = queryUser.FriendList
	} else if request.FORg == models.Group {
		stringList = queryUser.GroupList
	} else {
		return nil, common.MakeNormalErr(definition.InvalidFOrg.Code, definition.InvalidFOrg.Msg+request.FORg).Err()
	}
	//	4. 列表遍历
	list := strings.Split(stringList, ",")
	for _, id := range list {
		if target == id {
			return &UserRelationDecideResponse{IsRelated: true}, nil
		}
	}

	return &UserRelationDecideResponse{IsRelated: false}, nil
}
