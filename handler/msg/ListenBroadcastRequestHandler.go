package msg

import (
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"

	"github.com/tursom-im/context"
	"github.com/tursom-im/handler"
	"github.com/tursom-im/im_conn"
	"github.com/tursom-im/tursom_im_protobuf"
)

type listenBroadcastRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterImHandlerFactory(func(ctx *context.GlobalContext) handler.ImMsgHandler {
		return &listenBroadcastRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *listenBroadcastRequestHandler) HandleMsg(conn *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg, ctx util.ContextMap) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}
	if _, ok = msg.GetContent().(*tursom_im_protobuf.ImMsg_ListenBroadcastRequest); !ok {
		return
	}

	listenBroadcastRequest := msg.GetListenBroadcastRequest()
	response := &tursom_im_protobuf.ListenBroadcastResponse{
		ReqId: listenBroadcastRequest.ReqId,
	}
	handler.ResponseCtxKey.Get(ctx).Content = &tursom_im_protobuf.ImMsg_ListenBroadcastResponse{
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
