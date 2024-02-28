package vo

var UserVoHelper UserVoHelper_

type UserVoHelper_ struct{}

type UserVo struct {
	RegisterVo        RegisterVo
	LoginMobileVo     LoginMobileVo
	LoginPasswdVo     LoginPasswdVo
	BindMobileVo      BindMobileVo
	RepwdCheckVo      RepwdCheckVo
	RepwdVo           RepwdVo
	UpdateUserInfoVo  UpdateUserInfoVo
	PositionVo        PositionVo
	PosXYR            PosXYR
	CreateGroupVo     CreateGroupVo
	UpdateGroupInfoVo UpdateGroupInfoVo
	DeleteUserVo      DeleteUserVo
	ExitGroupVo       ExitGroupVo
	ToBeAdminVo       ToBeAdminVo
}

func (*UserVoHelper_) NewUserVo() *UserVo {
	return &UserVo{}
}

type RegisterVo struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Repassword string `json:"repassword"`
}

type LoginMobileVo struct {
	Mobile  string `json:"mobile"`
	Captcha string `json:"captcha"`
}

type LoginPasswdVo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BindMobileVo struct {
	Mobile  string `json:"mobile"`
	Captcha string `json:"captcha"`
}

type RepwdCheckVo struct {
	Username string `json:"username"`
	Mobile   string `json:"mobile"`
	Captcha  string `json:"captcha"`
}

type RepwdVo struct {
	Mobile     string `json:"mobile"`
	Password   string `json:"password"`
	Repassword string `json:"repassword"`
}

type UpdateUserInfoVo struct {
	Username string `json:"username"`
	Sex      string `json:"sex"`
	Sign     string `json:"sign"`
}

type PositionVo struct {
	X float64 `json:"x"` //	经度
	Y float64 `json:"y"` //	纬度
}

type PosXYR struct {
	X      float64 `json:"x"`      //	经度
	Y      float64 `json:"y"`      //	纬度
	R      int     `json:"r"`      //	半径
	Offset int     `json:"offset"` //	偏移量
}

type CreateGroupVo struct {
	GroupName    string `json:"group_name"`     //	群名
	GroupPost    string `json:"group_post"`     //	群公告
	MaxPersonNum int    `json:"max_person_num"` //	最大人数
}

type UpdateGroupInfoVo struct {
	Id           int    `json:"id"`
	GroupName    string `json:"group_name"`
	GroupPost    string `json:"group_post"`
	MaxPersonNum int    `json:"max_person_num"`
}

type DeleteUserVo struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
}

type ExitGroupVo struct {
	GroupId uint32 `json:"group_id"`
}

type ToBeAdminVo struct {
	UserId  uint32 `json:"user_id"`
	GroupId uint32 `json:"group_id"`
}
