package msg

import (
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"
	"google.golang.org/protobuf/proto"

	"github.com/tursom-im/context"
	"github.com/tursom-im/handler"
	"github.com/tursom-im/im_conn"
	"github.com/tursom-im/tursom_im_protobuf"
)

type sendBroadcastRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterImHandlerFactory(func(ctx *context.GlobalContext) handler.ImMsgHandler {
		return &sendBroadcastRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *sendBroadcastRequestHandler) HandleMsg(conn *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg, ctx util.ContextMap) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}
	if _, ok = msg.GetContent().(*tursom_im_protobuf.ImMsg_SendBroadcastRequest); !ok {
		return
	}

	sendBroadcastRequest := msg.GetSendBroadcastRequest()
	response := &tursom_im_protobuf.SendBroadcastResponse{
		ReqId: sendBroadcastRequest.ReqId,
	}
	handler.ResponseCtxKey.Get(ctx).Content = &tursom_im_protobuf.ImMsg_SendBroadcastResponse{
		SendBroadcastResponse: response,
	}

	broadcast := &tursom_im_protobuf.ImMsg{
		MsgId: h.globalContext.MsgIdContext().NewMsgIdStr(),
		Content: &tursom_im_protobuf.ImMsg_Broadcast{Broadcast: &tursom_im_protobuf.Broadcast{
			Sender:  h.globalContext.AttrContext().UserIdAttrKey().Get(conn).Get().String(),
			ReqId:   sendBroadcastRequest.ReqId,
			Channel: sendBroadcastRequest.Channel,
			Content: sendBroadcastRequest.Content,
		}},
	}
	bytes, err := proto.Marshal(broadcast)
	if err != nil {
		exceptions.Print(err)
	} else {
		response.ReceiverCount = h.globalContext.BroadcastContext().Send(
			sendBroadcastRequest.Channel,
			bytes,
			func(c *im_conn.AttachmentConn) bool {
				return c != conn
			},
		)
	}
	return
}
