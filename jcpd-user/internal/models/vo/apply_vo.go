package vo

var ApplyVoHelper ApplyVoHelper_

type ApplyVoHelper_ struct{}

type JoinApplyVo struct {
	ApplyFriendVo       ApplyFriendVo
	UpdateApplyStatusVo UpdateApplyStatusVo
}

func (*ApplyVoHelper_) NewApplyVo() *JoinApplyVo {
	return &JoinApplyVo{}
}

type ApplyFriendVo struct {
	FriendName   string `json:"friend_name"`
	Introduction string `json:"introduction"`
}

type UpdateApplyStatusVo struct {
	Username  string `json:"username"`
	ApplyType string `json:"apply_type"`
	CurStatus string `json:"cur_status"`
	ToStatus  string `json:"to_status"`
}
