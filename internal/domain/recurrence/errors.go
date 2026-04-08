package recurrence

import "errors"

var (
	ErrNotFound      = errors.New("recurrence rule not found")
	ErrInvalidType   = errors.New("invalid recurrence type")
	ErrInvalidDates  = errors.New("invalid date range")
	ErrTooManyTasks  = errors.New("too many tasks would be generated (max 500)")
	ErrInvalidParams = errors.New("invalid recurrence parameters")
)
