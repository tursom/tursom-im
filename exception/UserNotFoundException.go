package exception

import "github.com/tursom/GoCollections/exceptions"

type UserNotFoundException struct {
	exceptions.RuntimeException
}

func NewUserNotFoundException(message interface{}) UserNotFoundException {
	return UserNotFoundException{
		exceptions.NewRuntimeException(
			message,
			"exception caused TokenSigException:",
			true,
			nil,
		),
	}
}
