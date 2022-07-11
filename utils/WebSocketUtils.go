package utils

import (
	"io"
	"net"

	"github.com/gobwas/ws/wsutil"
	"github.com/tursom/GoCollections/exceptions"
)

func IsClosedError(err any) bool {
	unpackErr := exceptions.UnpackException(err)
	for unpackErr != err {
		err = unpackErr
		unpackErr = exceptions.UnpackException(err)
	}
	if err == nil {
		return false
	}
	_, ok := err.(wsutil.ClosedError)
	ok = ok || err == io.EOF
	{
		_, opError := err.(net.OpError)
		ok = ok || opError
	}
	return ok
}
