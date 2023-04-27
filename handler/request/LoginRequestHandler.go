package request

import (
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"

	"github.com/tursom/tursom-im/conn"
	"github.com/tursom/tursom-im/context"
	"github.com/tursom/tursom-im/exception"
	"github.com/tursom/tursom-im/handler"
	m "github.com/tursom/tursom-im/proto/msg"
	"github.com/tursom/tursom-im/proto/msys"
)

type loginRequestHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func init() {
	handler.RegisterMsgHandlerFactory(func(ctx *context.GlobalContext) handler.IMMsgHandler {
		return &loginRequestHandler{
			globalContext: ctx,
		}
	})
}

func (h *loginRequestHandler) HandleMsg(c conn.Conn, msg *m.ImMsg, ctx util.ContextMap) (ok bool) {
	if h == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", nil))
	}

	if _, ok = msg.GetContent().(*m.ImMsg_LoginRequest); !ok {
		return
	}

	loginResult := &m.LoginResult{}
	handler.ResponseCtxKey.Get(ctx).Content = &m.ImMsg_LoginResult{LoginResult: loginResult}

	token, err := h.globalContext.Token().Parse(msg.GetLoginRequest().Token)
	if err != nil {
		exception.Log("handler/msg/LoginRequestHandler.go: an exception caused on parse token", err)
		return
	}

	uid := token.Uid
	if msg.GetLoginRequest().GetTempId() {
		uid = uid + "-" + h.globalContext.MsgId().NewMsgIdStr()
	}

	h.globalContext.Attr().UserId(c).Set(lang.NewString(uid))
	h.globalContext.Attr().UserToken(c).Set(token)

	h.globalContext.Broadcast().Listen(msys.UserChannel(uid), c)
	c.AddEventListener(func(event conn.Event) {
		if !event.EventId().IsConnClosed() {
			return
		}
		h.globalContext.Broadcast().CancelListen(msys.UserChannel(uid), c)
	})

	loginResult.ImUserId = uid
	loginResult.Success = true
	return
}
