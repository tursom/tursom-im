package handler

import (
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"

	"github.com/tursom-im/conn"
	"github.com/tursom-im/context"
	"github.com/tursom-im/proto/pkg"
)

var (
	logicHandlerFactories []func(ctx *context.GlobalContext) IMLogicHandler
	imMsgContext          = util.NewContext()
	// ResponseCtxKey if HandleMsg returns true, then this value must not be nil
	ResponseCtxKey = util.AllocateContextKeyWithDefault[*pkg.ImMsg](
		imMsgContext,
		func() *pkg.ImMsg {
			return &pkg.ImMsg{}
		},
	)
	// CloseConnectionCtxKey if this value on ctx is true, then close this connection
	CloseConnectionCtxKey = util.AllocateContextKey[bool](imMsgContext)
)

type (
	// IMLogicHandler shows an object that can handle im msg
	// default im handlers on package handler/msg, you need to run msg.Init() to initial logicHandlerFactories
	IMLogicHandler interface {
		lang.Object
		// HandleMsg
		// ***WARN***: if returns true, value of ResponseCtxKey on ctx must be set and not be nil
		HandleMsg(
			c conn.Conn,
			msg *pkg.ImMsg,
			ctx util.ContextMap,
		) (ok bool)
	}
)

func RegisterLogicHandlerFactory(handlerFactory func(ctx *context.GlobalContext) IMLogicHandler) {
	logicHandlerFactories = append(logicHandlerFactories, handlerFactory)
}

func LogicHandlers(ctx *context.GlobalContext) []IMLogicHandler {
	handlers := make([]IMLogicHandler, len(logicHandlerFactories))
	for i, factory := range logicHandlerFactories {
		handlers[i] = factory(ctx)
	}
	return handlers
}

func NewImMsgContext() util.ContextMap {
	return imMsgContext.NewMap()
}
