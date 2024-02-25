package dto

// ApplyInfoDto 这里不用id作为返回条件可能更有利于分库分表
type ApplyInfoDto struct {
	Username     string `json:"username"`
	Sex          string `json:"sex"`
	Status       string `json:"status"`
	Introduction string `json:"introduction"`
}

type ApplyInfoDtos []ApplyInfoDto
