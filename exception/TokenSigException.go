package exception

import "github.com/tursom/GoCollections/exceptions"

type TokenSigException struct {
	exceptions.RuntimeException
}

func NewTokenSigException(message any) TokenSigException {
	return TokenSigException{
		exceptions.NewRuntimeException(
			message,
			"exception caused TokenSigException:",
			nil,
		),
	}
}
