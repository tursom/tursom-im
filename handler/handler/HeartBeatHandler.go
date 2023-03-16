package handler

import (
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"

	"github.com/tursom-im/conn"
	"github.com/tursom-im/context"
	"github.com/tursom-im/handler"
	"github.com/tursom-im/proto/pkg"
)

type heartBeatHandler struct {
	lang.BaseObject
}

func init() {
	handler.RegisterLogicHandlerFactory(func(_ *context.GlobalContext) handler.IMLogicHandler {
		return &heartBeatHandler{}
	})
}

func (h *heartBeatHandler) HandleMsg(_ conn.Conn, msg *pkg.ImMsg, ctx util.ContextMap) (ok bool) {
	if _, ok = msg.GetContent().(*pkg.ImMsg_HeartBeat); !ok {
		return
	}
	handler.ResponseCtxKey.Get(ctx).Content = msg.Content
	return
}
