package exception

import "github.com/tursom/GoCollections/exceptions"

type TokenParseException struct {
	exceptions.RuntimeException
}

func NewTokenParseException(message string) TokenParseException {
	return TokenParseException{
		exceptions.NewRuntimeException(message, exceptions.DefaultExceptionConfig().AddSkipStack(1).
			SetExceptionName("github.com.tursom.tursom-im.exception.TokenParseException")),
	}
}
