package exception

import "github.com/tursom/GoCollections/exceptions"

type IllegalParameterException struct {
	exceptions.RuntimeException
}

func NewIllegalParameterException(message string, config *exceptions.ExceptionConfig) *IllegalParameterException {
	return &IllegalParameterException{
		exceptions.NewRuntimeException(message, config.AddSkipStack(1).
			SetExceptionName("IllegalParameterException")),
	}
}
