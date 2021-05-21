package exception

import "github.com/tursom/GoCollections/exceptions"

type InvalidTypeException struct {
	exceptions.RuntimeException
}

func NewInvalidTypeException(message interface{}) InvalidTypeException {
	return InvalidTypeException{
		exceptions.NewRuntimeException(
			message,
			"exception caused InvalidTypeException:",
			true,
			nil,
		),
	}
}
