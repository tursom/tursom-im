package msg

import (
	"github.com/tursom-im/context"
	"github.com/tursom-im/handler"
	"github.com/tursom-im/im_conn"
	"github.com/tursom-im/tursom_im_protobuf"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"
)

type allocateNodeRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterImHandlerFactory(func(ctx *context.GlobalContext) handler.ImMsgHandler {
		return &allocateNodeRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *allocateNodeRequestHandler) HandleMsg(conn *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg, ctx util.ContextMap) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}
	if _, ok = msg.GetContent().(*tursom_im_protobuf.ImMsg_AllocateNodeRequest); !ok {
		return
	}

	allocateNodeResponse := &tursom_im_protobuf.AllocateNodeResponse{
		ReqId: msg.GetAllocateNodeRequest().ReqId,
	}
	handler.ResponseCtxKey.Get(ctx).Content = &tursom_im_protobuf.ImMsg_AllocateNodeResponse{
		AllocateNodeResponse: allocateNodeResponse,
	}

	allocateNodeResponse.Node = h.globalContext.ConnNodeContext().Allocate(conn)
	return
}
