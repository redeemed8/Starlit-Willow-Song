package dto

type GroupInfoDto struct {
	Id           uint32   `json:"id"`
	GroupName    string   `json:"group_name"`
	GroupPost    string   `json:"group_post"`
	LordId       uint32   `json:"lord_id"`
	AdminIds     []uint32 `json:"admin_ids"`
	MemberIds    []uint32 `json:"member_ids"`
	CurPersonNum int      `json:"cur_person_num"`
	MaxPersonNum int      `json:"max_person_num"`
}
