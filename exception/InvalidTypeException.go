package exception

import "github.com/tursom/GoCollections/exceptions"

type InvalidTypeException struct {
	exceptions.RuntimeException
}

func NewInvalidTypeException(message string) InvalidTypeException {
	return InvalidTypeException{
		exceptions.NewRuntimeException(message, exceptions.DefaultExceptionConfig().AddSkipStack(1).
			SetExceptionName("InvalidTypeException")),
	}
}
