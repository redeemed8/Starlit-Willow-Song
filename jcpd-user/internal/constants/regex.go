package constants

const (
	UsernameRegex  = `^[\p{Han}a-zA-Z0-9]+$`
	SignRegex      = `^[\p{Han}a-zA-Z0-9,，.。?？!！]+$`
	GroupNameRegex = `^[\p{Han}a-zA-Z0-9,，.。?？！!-=_]+$`
	GroupPostRegex = `^[\p{Han}a-zA-Z0-9,，.。?？！!-=_]+$`

	IntroduceRegex = `[\p{Han}\p{N}\p{L}\p{P}]`
)
