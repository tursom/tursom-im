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

type allocateNodeRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterLogicHandlerFactory(func(ctx *context.GlobalContext) handler.IMLogicHandler {
		return &allocateNodeRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *allocateNodeRequestHandler) HandleMsg(
	conn conn.Conn,
	msg *pkg.ImMsg,
	ctx util.ContextMap,
) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}
	if _, ok = msg.GetContent().(*pkg.ImMsg_AllocateNodeRequest); !ok {
		return false
	}

	allocateNodeResponse := &pkg.AllocateNodeResponse{
		ReqId: msg.GetAllocateNodeRequest().ReqId,
	}
	handler.ResponseCtxKey.Get(ctx).Content = &pkg.ImMsg_AllocateNodeResponse{
		AllocateNodeResponse: allocateNodeResponse,
	}

	allocateNodeResponse.Node = h.globalContext.ConnNodeContext().Allocate(conn)
	return
}
