package errs

import "errors"

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrMaxTasksReached   = errors.New("server is busy (max tasks limit)")
	ErrMaxObjectsReached = errors.New("maximum objects per task reached")
	ErrInvalidFileType   = errors.New("invalid file type (allowed: .pdf, .jpeg)")
)
