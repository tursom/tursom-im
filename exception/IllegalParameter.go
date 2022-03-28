package exception

import "github.com/tursom/GoCollections/exceptions"

type IllegalParameterException struct {
	exceptions.RuntimeException
}

func NewIllegalParameterException(message any, config *exceptions.ExceptionConfig) *IllegalParameterException {
	return &IllegalParameterException{
		exceptions.NewRuntimeException(
			message,
			"exception caused ElementNotFoundException:",
			config.AddSkipStack(1),
		),
	}
}
