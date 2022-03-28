package exception

import "github.com/tursom/GoCollections/exceptions"

type TokenParseException struct {
	exceptions.RuntimeException
}

func NewTokenParseException(message any) TokenParseException {
	return TokenParseException{
		exceptions.NewRuntimeException(
			message,
			"exception caused TokenParseException:",
			nil,
		),
	}
}
