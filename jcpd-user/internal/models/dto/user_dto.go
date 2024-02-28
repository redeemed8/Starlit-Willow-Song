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

type UserInfoDto2 struct {
	Id       uint32 `json:"id"`
	Username string `json:"username"`
	Sex      string `json:"sex"`
	Sign     string `json:"sign"`
}

type NearbyUsers []NearbyUserDto

type NearbyUserDto struct {
	Username string `json:"username"`
	Sex      string `json:"sex"`
	Sign     string `json:"sign"`
	Distance string `json:"distance"`
}
