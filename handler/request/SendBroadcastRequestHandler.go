package request

import (
	"strconv"

	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"

	"github.com/tursom/tursom-im/conn"
	"github.com/tursom/tursom-im/context"
	"github.com/tursom/tursom-im/handler"
	m "github.com/tursom/tursom-im/proto/msg"
	"github.com/tursom/tursom-im/proto/msys"
)

type sendBroadcastRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterMsgHandlerFactory(func(ctx *context.GlobalContext) handler.IMMsgHandler {
		return &sendBroadcastRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *sendBroadcastRequestHandler) HandleMsg(c conn.Conn, msg *m.ImMsg, ctx util.ContextMap) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}
	if _, ok = msg.GetContent().(*m.ImMsg_SendBroadcastRequest); !ok {
		return
	}

	sendBroadcastRequest := msg.GetSendBroadcastRequest()
	response := &m.SendBroadcastResponse{
		ReqId: sendBroadcastRequest.ReqId,
	}
	handler.ResponseCtxKey.Get(ctx).Content = &m.ImMsg_SendBroadcastResponse{
		SendBroadcastResponse: response,
	}

	broadcast := &m.ImMsg{
		MsgId: h.globalContext.MsgId().NewMsgIdStr(),
		Content: &m.ImMsg_Broadcast{Broadcast: &m.Broadcast{
			Sender:  h.globalContext.Attr().UserId(c).Get().String(),
			ReqId:   sendBroadcastRequest.ReqId,
			Channel: sendBroadcastRequest.Channel,
			Content: sendBroadcastRequest.Content,
		}},
	}

	h.globalContext.Broadcast().Send(
		uint32(msys.Channel_GROUP),
		strconv.Itoa(int(sendBroadcastRequest.Channel)),
		broadcast,
	)
	response.ReceiverCount = -1

	return
}
