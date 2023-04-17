package handler

import (
	"sync"

	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"

	"github.com/tursom-im/conn"
	"github.com/tursom-im/context"
	m "github.com/tursom-im/proto/msg"
	"github.com/tursom-im/proto/msys"
)

var (
	factoriesLock       sync.Mutex
	msgHandlerFactories []func(ctx *context.GlobalContext) IMMsgHandler
	imMsgContext        = util.NewContext()
	// ResponseCtxKey if HandleMsg returns true, then this value must not be nil
	ResponseCtxKey = util.AllocateContextKeyWithDefault(
		imMsgContext,
		func() *m.ImMsg {
			return &m.ImMsg{}
		},
	)
	// CloseConnectionCtxKey if this value on ctx is true, then close this connection
	CloseConnectionCtxKey = util.AllocateContextKey[bool](imMsgContext)
)

type (
	// IMMsgHandler shows an object that can handle im msg
	// default im handlers on package handler/msg, you need to run msg.Init() to initial msgHandlerFactories
	IMMsgHandler interface {
		lang.Object
		// HandleMsg
		// ***WARN***: if returns true, value of ResponseCtxKey on ctx must be set and not be nil
		HandleMsg(
			c conn.Conn,
			msg *m.ImMsg,
			ctx util.ContextMap,
		) (ok bool)
	}

	SystemMsgHandler interface {
		lang.Object
		HandleSystemMsg(
			msg msys.SystemMsg,
			ctx util.ContextMap,
		) (ok bool)
	}
)

func RegisterMsgHandlerFactory(handlerFactory func(ctx *context.GlobalContext) IMMsgHandler) {
	factoriesLock.Lock()
	defer factoriesLock.Unlock()

	msgHandlerFactories = append(msgHandlerFactories, handlerFactory)
}

func MsgHandlers(ctx *context.GlobalContext) []IMMsgHandler {
	factoriesLock.Lock()
	defer factoriesLock.Unlock()

	handlers := make([]IMMsgHandler, len(msgHandlerFactories))
	for i, factory := range msgHandlerFactories {
		handlers[i] = factory(ctx)
	}
	return handlers
}

func NewImMsgContext() util.ContextMap {
	return imMsgContext.NewMap()
}
