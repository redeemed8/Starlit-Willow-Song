package vo

var ApplyVoHelper ApplyVoHelper_

type ApplyVoHelper_ struct{}

type JoinApplyVo struct {
	ApplyFriendVo ApplyFriendVo
}

func (*ApplyVoHelper_) NewApplyVo() *JoinApplyVo {
	return &JoinApplyVo{}
}

type ApplyFriendVo struct {
	FriendName   string `json:"friend_name"`
	Introduction string `json:"introduction"`
}
