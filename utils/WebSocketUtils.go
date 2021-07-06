package utils

import (
	"github.com/gobwas/ws/wsutil"
	"github.com/tursom/GoCollections/exceptions"
	"io"
)

func IsClosedError(err interface{}) bool {
	err = exceptions.UnpackException(err)
	_, ok := err.(wsutil.ClosedError)
	ok = ok || err == io.EOF
	return ok
}
