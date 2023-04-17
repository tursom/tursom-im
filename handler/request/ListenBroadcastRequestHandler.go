package request

import (
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"

	"github.com/tursom-im/conn"
	"github.com/tursom-im/context"
	"github.com/tursom-im/handler"
	m "github.com/tursom-im/proto/msg"
)

type listenBroadcastRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterMsgHandlerFactory(func(ctx *context.GlobalContext) handler.IMMsgHandler {
		return &listenBroadcastRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *listenBroadcastRequestHandler) HandleMsg(conn conn.Conn, msg *m.ImMsg, ctx util.ContextMap) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}
	if _, ok = msg.GetContent().(*m.ImMsg_ListenBroadcastRequest); !ok {
		return
	}

	listenBroadcastRequest := msg.GetListenBroadcastRequest()
	response := &m.ListenBroadcastResponse{
		ReqId: listenBroadcastRequest.ReqId,
	}
	handler.ResponseCtxKey.Get(ctx).Content = &m.ImMsg_ListenBroadcastResponse{
		ListenBroadcastResponse: response,
	}

	var err exceptions.Exception = nil
	if listenBroadcastRequest.CancelListen {
		err = h.globalContext.Broadcast().CancelListen(listenBroadcastRequest.Channel, conn)
	} else {
		err = h.globalContext.Broadcast().Listen(listenBroadcastRequest.Channel, conn)
	}
	if err != nil {
		err.PrintStackTrace()
	} else {
		response.Success = true
	}
	return
}
