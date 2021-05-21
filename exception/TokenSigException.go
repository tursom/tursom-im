package exception

import "github.com/tursom/GoCollections/exceptions"

type TokenSigException struct {
	exceptions.RuntimeException
}

func NewTokenSigException(message interface{}) TokenSigException {
	return TokenSigException{
		exceptions.NewRuntimeException(
			message,
			"exception caused TokenSigException:",
			true,
			nil,
		),
	}
}
