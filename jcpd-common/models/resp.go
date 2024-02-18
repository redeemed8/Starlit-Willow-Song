package common_models

type StatusCode int
type Resp struct {
	Code StatusCode `json:"code"`
	Data any        `json:"data"`
	Msg  string     `json:"msg"`
}

func NewResp() *Resp {
	return &Resp{}
}

const OK = 2000

func (r *Resp) Success(data any) *Resp {
	r.Code = OK
	r.Msg = "success"
	r.Data = data
	return r
}

func (r *Resp) Fail(err NormalErr) *Resp {
	r.Code = StatusCode(err.Code)
	r.Msg = err.Msg
	return r
}
