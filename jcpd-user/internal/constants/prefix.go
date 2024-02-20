package constants

import "time"

const (
	RdbLoginCode  = "captcha.login."
	RdbForgetCode = "captcha.forget."
	RdbBindCode   = "captcha.bind."
)

var captchaModeMap = newCM()

const (
	LoginMode  = "0"
	ForgetMode = "1"
	BindMode   = "2"
)

func newCM() map[string]string {
	map_ := make(map[string]string)
	map_[LoginMode] = RdbLoginCode
	map_[ForgetMode] = RdbForgetCode
	map_[BindMode] = RdbBindCode
	return map_
}

func MatchModeCode(mode string) (prefix string) {
	prefix = captchaModeMap[mode]
	return prefix
}

const (
	RepwdCheckPrefix = "repwd.check."
	RepwdCheckExpire = 10 * time.Minute
)
