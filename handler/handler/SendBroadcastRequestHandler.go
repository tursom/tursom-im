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

type sendBroadcastRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterLogicHandlerFactory(func(ctx *context.GlobalContext) handler.IMLogicHandler {
		return &sendBroadcastRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *sendBroadcastRequestHandler) HandleMsg(c conn.Conn, msg *pkg.ImMsg, ctx util.ContextMap) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}
	if _, ok = msg.GetContent().(*pkg.ImMsg_SendBroadcastRequest); !ok {
		return
	}

	sendBroadcastRequest := msg.GetSendBroadcastRequest()
	response := &pkg.SendBroadcastResponse{
		ReqId: sendBroadcastRequest.ReqId,
	}
	handler.ResponseCtxKey.Get(ctx).Content = &pkg.ImMsg_SendBroadcastResponse{
		SendBroadcastResponse: response,
	}

	broadcast := &pkg.ImMsg{
		MsgId: h.globalContext.MsgIdContext().NewMsgIdStr(),
		Content: &pkg.ImMsg_Broadcast{Broadcast: &pkg.Broadcast{
			Sender:  h.globalContext.AttrContext().UserIdAttrKey().Get(c).Get().String(),
			ReqId:   sendBroadcastRequest.ReqId,
			Channel: sendBroadcastRequest.Channel,
			Content: sendBroadcastRequest.Content,
		}},
	}

	response.ReceiverCount = int32(h.globalContext.BroadcastContext().Send(
		sendBroadcastRequest.Channel,
		broadcast,
		func(c conn.Conn) bool {
			return c != c
		},
	))
	return
}
