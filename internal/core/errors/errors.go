package errors

import "errors"

var (
	ErrPreconditionFailed = errors.New("precondition failed")
	ErrActionFailed       = errors.New("action failed")
	ErrPartialSuccess     = errors.New("partial success")
	ErrLocked             = errors.New("another opskit operation is running")
)
