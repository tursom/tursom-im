package exception

import "github.com/tursom/GoCollections/exceptions"

type TokenParseException struct {
	exceptions.RuntimeException
}

func NewTokenParseException(message interface{}) TokenParseException {
	return TokenParseException{
		exceptions.NewRuntimeException(
			message,
			"exception caused TokenParseException:",
			true,
			nil,
		),
	}
}
