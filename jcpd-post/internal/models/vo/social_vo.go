package vo

// SocialVoHelper  满足 对外提供
var SocialVoHelper socialVoHelper_

type socialVoHelper_ struct{}

type socialVo struct {
	LikePostVo       likePostVo
	PublishCommentVo publishCommentVo
	DeleteCommentVo  deleteCommentVo
}

func (*socialVoHelper_) NewSocialVo() *socialVo {
	return &socialVo{}
}

type likePostVo struct {
	PostId uint32 `json:"post_id"`
}

type publishCommentVo struct {
	PostId  uint32 `json:"post_id"`
	Content string `json:"content"`
}

type deleteCommentVo struct {
	PostId uint32 `json:"post_id"`
}
