package vo

var UserVoHelper UserVoHelper_

type UserVoHelper_ struct{}

type UserVo struct {
	RegisterVo    RegisterVo
	LoginMobileVo LoginMobileVo
	BindMobileVo  BindMobileVo
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

type BindMobileVo struct {
	Mobile  string `json:"mobile"`
	Captcha string `json:"captcha"`
}
