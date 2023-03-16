package exception

import "github.com/tursom/GoCollections/exceptions"

type NodeOfflineException struct {
	exceptions.RuntimeException
}

func NewNodeOfflineException(message string) *NodeOfflineException {
	return &NodeOfflineException{
		*exceptions.NewRuntimeException(message, exceptions.DefaultExceptionConfig().AddSkipStack(1).
			SetExceptionName("github.com.tursom.tursom-im.exception.NodeOfflineException")),
	}
}
