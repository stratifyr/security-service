package services

var (
	ErrForbidden           = &Error{403, "forbidden"}
	ErrDateRangeTooLong    = &Error{400, "date range is too long"}
	ErrInvalidStatusChange = &Error{400, "invalid status change"}
	ErrMarketHolidayStat   = &Error{400, "cannot add stat for market holiday"}
)

type Error struct {
	httpCode int
	message  string
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) StatusCode() int {
	return e.httpCode
}
