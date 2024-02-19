package vo

var UserVoHelper UserVoHelper_

type UserVoHelper_ struct{}

type UserVo struct {
	RegisterVo       RegisterVo
	LoginMobileVo    LoginMobileVo
	LoginPasswdVo    LoginPasswdVo
	BindMobileVo     BindMobileVo
	RepwdCheckVo     RepwdCheckVo
	RepwdVo          RepwdVo
	UpdateUserInfoVo UpdateUserInfoVo
	PositionVo       PositionVo
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
	Password   string `json:"password"`
	Repassword string `json:"repassword"`
}

type UpdateUserInfoVo struct {
	Username string `json:"username"`
	Sex      string `json:"sex"`
	Sign     string `json:"sign"`
}

type PositionVo struct {
	Longitude float64 `json:"longitude"` //	经度
	Latitude  float64 `json:"latitude"`  //	纬度
}
