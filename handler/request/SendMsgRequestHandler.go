package request

import (
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"

	"github.com/tursom/tursom-im/conn"
	"github.com/tursom/tursom-im/context"
	"github.com/tursom/tursom-im/handler"
	m "github.com/tursom/tursom-im/proto/msg"
	"github.com/tursom/tursom-im/proto/msys"
)

type sendMsgRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterMsgHandlerFactory(func(ctx *context.GlobalContext) handler.IMMsgHandler {
		return &sendMsgRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *sendMsgRequestHandler) HandleMsg(c conn.Conn, msg *m.ImMsg, ctx util.ContextMap) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}

	if _, ok = msg.GetContent().(*m.ImMsg_SendMsgRequest); !ok {
		return
	}

	response := &m.SendMsgResponse{}
	msgId := h.globalContext.MsgId().NewMsgIdStr()
	{
		r := handler.ResponseCtxKey.Get(ctx)
		r.Content, r.MsgId = &m.ImMsg_SendMsgResponse{SendMsgResponse: response}, msgId
	}

	sendMsgRequest := msg.GetSendMsgRequest()
	sender := h.globalContext.Attr().UserId(c).Get().AsString()
	response.ReqId = sendMsgRequest.ReqId

	receiver := sendMsgRequest.Receiver

	response.Success = true
	imMsg := &m.ImMsg{
		MsgId: msgId,
		Content: &m.ImMsg_ChatMsg{ChatMsg: &m.ChatMsg{
			Receiver: receiver,
			Sender:   sender,
			Content:  sendMsgRequest.Content,
		}},
	}

	h.globalContext.Broadcast().Send(uint32(msys.Channel_USER), sender, imMsg)
	h.globalContext.Broadcast().Send(uint32(msys.Channel_USER), receiver, imMsg)

	return
}
