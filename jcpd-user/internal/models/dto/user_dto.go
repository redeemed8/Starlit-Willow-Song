package dto

type UserInfoDto struct {
	Username string `json:"username"`
	Sex      string `json:"sex"`
	Sign     string `json:"sign"`
}

type UserInfoDtos []UserInfoDto

func (dtos UserInfoDtos) First() UserInfoDto {
	if len(dtos) == 0 {
		return UserInfoDto{}
	}
	return dtos[0]
}
