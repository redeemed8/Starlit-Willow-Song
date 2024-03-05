package vo

// PostVoHelper  满足 对外提供
var PostVoHelper postVoHelper_

type postVoHelper_ struct{}

type postVo struct {
	PublishPostVo publishPostVo
}

func (*postVoHelper_) NewPostVo() *postVo {
	return &postVo{}
}

type publishPostVo struct {
	Title    string `json:"title"`
	TopicTag string `json:"topicTag"`
	Body     string `json:"body"`
}
