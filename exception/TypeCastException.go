package exception

import (
	"fmt"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"reflect"
)

type TypeCastException struct {
	exceptions.RuntimeException
}

func NewTypeCastException(message any, config *exceptions.ExceptionConfig) *TypeCastException {
	return &TypeCastException{
		exceptions.NewRuntimeException(
			message,
			"exception caused ElementNotFoundException:",
			config.AddSkipStack(1),
		),
	}
}

func NewTypeCastExceptionByType[T any](obj any, config *exceptions.ExceptionConfig) *TypeCastException {
	return NewTypeCastException(
		fmt.Sprintf("object %s cannot cast to %s", obj, reflect.TypeOf(lang.Nil[T]()).Name()),
		config,
	)
}
