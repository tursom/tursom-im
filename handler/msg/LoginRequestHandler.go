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

type loginRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterImHandlerFactory(func(ctx *context.GlobalContext) handler.ImMsgHandler {
		return &loginRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *loginRequestHandler) HandleMsg(conn *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg, ctx util.ContextMap) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}

	if _, ok = msg.GetContent().(*tursom_im_protobuf.ImMsg_LoginRequest); !ok {
		return
	}

	loginResult := &tursom_im_protobuf.LoginResult{}
	handler.ResponseCtxKey.Get(ctx).Content = &tursom_im_protobuf.ImMsg_LoginResult{LoginResult: loginResult}

	token, err := h.globalContext.TokenContext().Parse(msg.GetLoginRequest().Token)
	if err != nil {
		exceptions.Print(err)
		return
	}

	userIdAttr := h.globalContext.AttrContext().UserIdAttrKey().Get(conn)
	userTokenAttr := h.globalContext.AttrContext().UserTokenAttrKey().Get(conn)

	uid := token.Uid
	if msg.GetLoginRequest().TempId {
		uid = uid + "-" + h.globalContext.MsgIdContext().NewMsgIdStr()
	}

	userIdAttr.Set(lang.NewString(uid))
	userTokenAttr.Set(token)

	connGroup := h.globalContext.UserConnContext().TouchUserConn(uid)
	connGroup.Add(conn)
	conn.AddEventListener(func(event im_conn.ConnEvent) {
		if !event.EventId().IsConnClosed() || connGroup.Size() != 0 {
			return
		}
		h.globalContext.UserConnContext().RemoveUserConn(uid)
	})

	loginResult.ImUserId = uid
	loginResult.Success = true
	return
}
