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

type allocateNodeRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterMsgHandlerFactory(func(ctx *context.GlobalContext) handler.IMMsgHandler {
		return &allocateNodeRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *allocateNodeRequestHandler) HandleMsg(
	conn conn.Conn,
	msg *m.ImMsg,
	ctx util.ContextMap,
) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}
	if _, ok = msg.GetContent().(*m.ImMsg_AllocateNodeRequest); !ok {
		return false
	}

	allocateNodeResponse := &m.AllocateNodeResponse{
		ReqId: msg.GetAllocateNodeRequest().ReqId,
	}
	handler.ResponseCtxKey.Get(ctx).Content = &m.ImMsg_AllocateNodeResponse{
		AllocateNodeResponse: allocateNodeResponse,
	}

	allocateNodeResponse.Node = h.globalContext.ConnNode().Allocate(conn)
	return
}
