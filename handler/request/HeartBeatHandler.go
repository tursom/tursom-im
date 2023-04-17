package request

import (
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"

	"github.com/tursom-im/conn"
	"github.com/tursom-im/context"
	"github.com/tursom-im/handler"
	m "github.com/tursom-im/proto/msg"
)

type heartBeatHandler struct {
	lang.BaseObject
}

func init() {
	handler.RegisterMsgHandlerFactory(func(_ *context.GlobalContext) handler.IMMsgHandler {
		return &heartBeatHandler{}
	})
}

func (h *heartBeatHandler) HandleMsg(_ conn.Conn, msg *m.ImMsg, ctx util.ContextMap) (ok bool) {
	if _, ok = msg.GetContent().(*m.ImMsg_HeartBeat); !ok {
		return
	}
	handler.ResponseCtxKey.Get(ctx).Content = msg.Content
	return
}
