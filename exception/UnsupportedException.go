package exception

import "github.com/tursom/GoCollections/exceptions"

type UnsupportedException struct {
	exceptions.RuntimeException
}

func NewUnsupportedException(message string) UnsupportedException {
	return UnsupportedException{
		exceptions.NewRuntimeException(message, exceptions.DefaultExceptionConfig().AddSkipStack(1).
			SetExceptionName("UnsupportedException")),
	}
}
