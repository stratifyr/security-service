package services

type ErrResp struct {
	Code    int
	Message string
}

func (e *ErrResp) Error() string {
	return e.Message
}

func (e *ErrResp) StatusCode() int {
	return e.Code
}
