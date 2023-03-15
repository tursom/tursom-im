package exception

import "github.com/tursom/GoCollections/exceptions"

type UserNotFoundException struct {
	exceptions.RuntimeException
}

func NewUserNotFoundException(message string) *UserNotFoundException {
	return &UserNotFoundException{
		*exceptions.NewRuntimeException(message, exceptions.DefaultExceptionConfig().AddSkipStack(1).
			SetExceptionName("github.com.tursom.tursom-im.exception.UserNotFoundException")),
	}
}
