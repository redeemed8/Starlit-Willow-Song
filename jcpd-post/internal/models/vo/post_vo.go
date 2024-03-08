package vo

// PostVoHelper  满足 对外提供
var PostVoHelper postVoHelper_

type postVoHelper_ struct{}

type postVo struct {
	PublishPostVo publishPostVo
	UpdatePostVo  updatePostVo
	DeletePostVo  deletePostVo
}

func (*postVoHelper_) NewPostVo() *postVo {
	return &postVo{}
}

type publishPostVo struct {
	Title    string `json:"title"`
	TopicTag string `json:"topic"`
	Body     string `json:"body"`
}

type updatePostVo struct {
	Title    string `json:"title"`
	TopicTag string `json:"topic"`
	Body     string `json:"body"`
}

type deletePostVo struct {
	PostId uint32 `json:"post_id"`
}

type reviewPostVo struct {
	PostId    uint32 `json:"post_id"`
	CurStatus string `json:"cur_status"`
	ToStatus  string `json:"to_status"`
}
