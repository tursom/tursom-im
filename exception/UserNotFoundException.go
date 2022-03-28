package exception

import "github.com/tursom/GoCollections/exceptions"

type UserNotFoundException struct {
	exceptions.RuntimeException
}

func NewUserNotFoundException(message any) UserNotFoundException {
	return UserNotFoundException{
		exceptions.NewRuntimeException(
			message,
			"exception caused TokenSigException:",
			nil,
		),
	}
}
