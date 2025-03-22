package types

type LoginReq struct {
	Mobile string `json:"mobile" validate:"mobile"`
	IDCard string `json:"id_card" validate:"idcard"`
}

func (r *LoginReq) Validate() error {
	return validate.Struct(r)
}
