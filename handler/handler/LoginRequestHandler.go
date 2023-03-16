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

type loginRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterLogicHandlerFactory(func(ctx *context.GlobalContext) handler.IMLogicHandler {
		return &loginRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *loginRequestHandler) HandleMsg(c conn.Conn, msg *pkg.ImMsg, ctx util.ContextMap) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}

	if _, ok = msg.GetContent().(*pkg.ImMsg_LoginRequest); !ok {
		return
	}

	loginResult := &pkg.LoginResult{}
	handler.ResponseCtxKey.Get(ctx).Content = &pkg.ImMsg_LoginResult{LoginResult: loginResult}

	token, err := h.globalContext.TokenContext().Parse(msg.GetLoginRequest().Token)
	if err != nil {
		exceptions.Print(err)
		return
	}

	userIdAttr := h.globalContext.AttrContext().UserIdAttrKey().Get(c)
	userTokenAttr := h.globalContext.AttrContext().UserTokenAttrKey().Get(c)

	uid := token.Uid
	if msg.GetLoginRequest().GetTempId() {
		uid = uid + "-" + h.globalContext.MsgIdContext().NewMsgIdStr()
	}

	userIdAttr.Set(lang.NewString(uid))
	userTokenAttr.Set(token)

	connGroup := h.globalContext.UserConnContext().TouchUserConn(uid)
	connGroup.Add(c)
	c.AddEventListener(func(event conn.Event) {
		if !event.EventId().IsConnClosed() || connGroup.Size() != 0 {
			return
		}
		h.globalContext.UserConnContext().RemoveUserConn(uid)
	})

	loginResult.ImUserId = uid
	loginResult.Success = true
	return
}
