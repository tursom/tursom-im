package exception

import "github.com/tursom/GoCollections/exceptions"

type UnsupportedException struct {
	exceptions.RuntimeException
}

func NewUnsupportedException(message any) UnsupportedException {
	return UnsupportedException{
		exceptions.NewRuntimeException(
			message,
			"exception caused UnsupportedException:",
			nil,
		),
	}
}
