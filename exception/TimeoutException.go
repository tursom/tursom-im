package exception

import "github.com/tursom/GoCollections/exceptions"

type TimeoutException struct {
	exceptions.RuntimeException
}

func NewTimeoutException(message string) *TimeoutException {
	return &TimeoutException{
		*exceptions.NewRuntimeException(message, exceptions.DefaultExceptionConfig().AddSkipStack(1).
			SetExceptionName("github.com.tursom.tursom-im.exception.TimeoutException")),
	}
}
