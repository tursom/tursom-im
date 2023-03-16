package handler

import (
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"

	"github.com/tursom-im/conn"
	"github.com/tursom-im/context"
	"github.com/tursom-im/handler"
	"github.com/tursom-im/proto/pkg"
)

type listenBroadcastRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterLogicHandlerFactory(func(ctx *context.GlobalContext) handler.IMLogicHandler {
		return &listenBroadcastRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *listenBroadcastRequestHandler) HandleMsg(conn conn.Conn, msg *pkg.ImMsg, ctx util.ContextMap) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}
	if _, ok = msg.GetContent().(*pkg.ImMsg_ListenBroadcastRequest); !ok {
		return
	}

	listenBroadcastRequest := msg.GetListenBroadcastRequest()
	response := &pkg.ListenBroadcastResponse{
		ReqId: listenBroadcastRequest.ReqId,
	}
	handler.ResponseCtxKey.Get(ctx).Content = &pkg.ImMsg_ListenBroadcastResponse{
		ListenBroadcastResponse: response,
	}

	var err exceptions.Exception = nil
	if listenBroadcastRequest.CancelListen {
		err = h.globalContext.BroadcastContext().CancelListen(listenBroadcastRequest.Channel, conn)
	} else {
		err = h.globalContext.BroadcastContext().Listen(listenBroadcastRequest.Channel, conn)
	}
	if err != nil {
		err.PrintStackTrace()
	} else {
		response.Success = true
	}
	return
}
