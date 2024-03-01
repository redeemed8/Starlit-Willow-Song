package vo

var ApplyVoHelper ApplyVoHelper_

type ApplyVoHelper_ struct{}

type JoinApplyVo struct {
	ApplyFriendVo         ApplyFriendVo
	UpdateApplyStatusVo   UpdateApplyStatusVo
	ApplyGroupVo          ApplyGroupVo
	UdtApplyGroupStatusVo UdtApplyGroupStatusVo
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
	CurStatus string `json:"cur_status"`
	ToStatus  string `json:"to_status"`
}

type ApplyGroupVo struct {
	GroupId      uint32 `json:"group_id"`
	Introduction string `json:"introduction"`
}

type UdtApplyGroupStatusVo struct {
	Username  string `json:"username"`
	GroupId   uint32 `json:"group_id"`
	CurStatus string `json:"cur_status"`
	ToStatus  string `json:"to_status"`
}
