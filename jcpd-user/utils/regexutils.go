package utils

import "regexp"

func VerifyMobile(mobile string) bool {
	if mobile == "" || len(mobile) != 11 {
		return false
	}
	regular := "^(13[0-9]|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18[0-9]|19[0-35-9])\\d{8}$"
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobile)
}
