package usctl

import "errors"

var (
	// praser error
	ErrFileNotFound = errors.New("file not found")
	ErrInvalidMeta  = errors.New("invalid meta type or meta type not found")
	ErrInvalidData  = errors.New("invalid data")
	ErrRegiterKind  = errors.New("register kind error")
	ErrSerialized   = errors.New("serialized error")

	// http error
	ErrNewRequest = errors.New("failed to create new request")
)
