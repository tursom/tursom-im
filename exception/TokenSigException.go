package exception

import "github.com/tursom/GoCollections/exceptions"

type TokenSigException struct {
	exceptions.RuntimeException
}

func NewTokenSigException(message string) *TokenSigException {
	return &TokenSigException{
		*exceptions.NewRuntimeException(message, exceptions.DefaultExceptionConfig().AddSkipStack(1).
			SetExceptionName("github.com.tursom.tursom-im.exception.TokenSigException")),
	}
}
