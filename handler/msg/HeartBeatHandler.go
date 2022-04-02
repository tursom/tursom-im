package msg

import (
	"github.com/tursom-im/context"
	"github.com/tursom-im/handler"
	"github.com/tursom-im/im_conn"
	"github.com/tursom-im/tursom_im_protobuf"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"
)

type heartBeatHandler struct {
	lang.BaseObject
}

func init() {
	handler.RegisterImHandlerFactory(func(_ *context.GlobalContext) handler.ImMsgHandler {
		return &heartBeatHandler{}
	})
}

func (h *heartBeatHandler) HandleMsg(_ *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg, ctx util.ContextMap) (ok bool) {
	if _, ok = msg.GetContent().(*tursom_im_protobuf.ImMsg_HeartBeat); !ok {
		return
	}
	handler.ResponseCtxKey.Get(ctx).Content = msg.Content
	return
}
