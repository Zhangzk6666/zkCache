package response

type resErr struct {
	Code int
	Msg  string
}

func (c *resErr) Error() string { // 实现接口
	return c.Msg
}

func NewErr(code int) error {
	return &resErr{
		Code: code,
		Msg:  getMsg(code).(string),
	}
}

func NewErrWithMsg(code int, msg string) error {
	return &resErr{
		Code: code,
		Msg:  msg,
	}
}
