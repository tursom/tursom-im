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

type sendMsgRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterLogicHandlerFactory(func(ctx *context.GlobalContext) handler.IMLogicHandler {
		return &sendMsgRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *sendMsgRequestHandler) HandleMsg(c conn.Conn, msg *pkg.ImMsg, ctx util.ContextMap) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}

	if _, ok = msg.GetContent().(*pkg.ImMsg_SendMsgRequest); !ok {
		return
	}

	response := &pkg.SendMsgResponse{}
	msgId := h.globalContext.MsgIdContext().NewMsgIdStr()
	{
		r := handler.ResponseCtxKey.Get(ctx)
		r.Content, r.MsgId = &pkg.ImMsg_SendMsgResponse{SendMsgResponse: response}, msgId
	}

	sendMsgRequest := msg.GetSendMsgRequest()
	sender := h.globalContext.AttrContext().UserIdAttrKey().Get(c).Get().AsString()
	response.ReqId = sendMsgRequest.ReqId

	receiver := sendMsgRequest.Receiver
	receiverConn := h.globalContext.UserConnContext().GetUserConn(receiver)
	currentConn := h.globalContext.UserConnContext().GetUserConn(sender)
	if receiverConn == nil || currentConn == nil {
		response.FailMsg = "user \"" + receiver + "\" not login"
		response.FailType = pkg.FailType_TARGET_NOT_LOGIN
		return
	}

	response.Success = true
	imMsg := &pkg.ImMsg{
		MsgId: msgId,
		Content: &pkg.ImMsg_ChatMsg{ChatMsg: &pkg.ChatMsg{
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
