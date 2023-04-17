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
	receiverConn := h.globalContext.UserConn().GetUserConn(receiver)
	currentConn := h.globalContext.UserConn().GetUserConn(sender)
	if receiverConn == nil || currentConn == nil {
		response.FailMsg = "user \"" + receiver + "\" not login"
		response.FailType = m.FailType_TARGET_NOT_LOGIN
		return
	}

	response.Success = true
	imMsg := &m.ImMsg{
		MsgId: msgId,
		Content: &m.ImMsg_ChatMsg{ChatMsg: &m.ChatMsg{
			Receiver: receiver,
			Sender:   sender,
			Content:  sendMsgRequest.Content,
		}},
	}
	currentConn.Aggregation(receiverConn).WriteChatMsg(imMsg, func(c conn.Conn) bool {
		return c != c
	})

	return
}
