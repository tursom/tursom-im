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

func NewTypeCastException(message string, config *exceptions.ExceptionConfig) *TypeCastException {
	return &TypeCastException{
		exceptions.NewRuntimeException(message, config.AddSkipStack(1).
			SetExceptionName("TypeCastException")),
	}
}

func NewTypeCastExceptionByType[T any](obj any, config *exceptions.ExceptionConfig) *TypeCastException {
	return NewTypeCastException(
		fmt.Sprintf("object %s cannot cast to %s", obj, reflect.TypeOf(lang.Nil[T]()).Name()),
		config,
	)
}
