package service

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"jcpd.cn/user/internal/models"
	"jcpd.cn/user/pkg/definition"
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
