package exception

import "github.com/tursom/GoCollections/exceptions"

type UnsupportedException struct {
	exceptions.RuntimeException
}

func NewUnsupportedException(message interface{}) UnsupportedException {
	return UnsupportedException{
		exceptions.NewRuntimeException(
			message,
			"exception caused UnsupportedException:",
			true,
			nil,
		),
	}
}
