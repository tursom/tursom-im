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

type sendMsgRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterImHandlerFactory(func(ctx *context.GlobalContext) handler.ImMsgHandler {
		return &sendMsgRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *sendMsgRequestHandler) HandleMsg(conn *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg, ctx util.ContextMap) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}

	if _, ok = msg.GetContent().(*tursom_im_protobuf.ImMsg_SendMsgRequest); !ok {
		return
	}

	response := &tursom_im_protobuf.SendMsgResponse{}
	msgId := h.globalContext.MsgIdContext().NewMsgIdStr()
	{
		r := handler.ResponseCtxKey.Get(ctx)
		r.Content, r.MsgId = &tursom_im_protobuf.ImMsg_SendMsgResponse{SendMsgResponse: response}, msgId
	}

	sendMsgRequest := msg.GetSendMsgRequest()
	sender := h.globalContext.AttrContext().UserIdAttrKey().Get(conn).Get().AsString()
	response.ReqId = sendMsgRequest.ReqId

	receiver := sendMsgRequest.Receiver
	receiverConn := h.globalContext.UserConnContext().GetUserConn(receiver)
	currentConn := h.globalContext.UserConnContext().GetUserConn(sender)
	if receiverConn == nil || currentConn == nil {
		response.FailMsg = "user \"" + receiver + "\" not login"
		response.FailType = tursom_im_protobuf.FailType_TARGET_NOT_LOGIN
		return
	}

	response.Success = true
	imMsg := &tursom_im_protobuf.ImMsg{
		MsgId: msgId,
		Content: &tursom_im_protobuf.ImMsg_ChatMsg{ChatMsg: &tursom_im_protobuf.ChatMsg{
			Receiver: receiver,
			Sender:   sender,
			Content:  sendMsgRequest.Content,
		}},
	}
	_ = currentConn.Aggregation(receiverConn).WriteChatMsg(imMsg, func(c *im_conn.AttachmentConn) bool {
		return conn != c
	})

	return
}
