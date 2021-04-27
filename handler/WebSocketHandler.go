package handler

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/protobuf/proto"
	"github.com/julienschmidt/httprouter"
	com_joinu_im_protobuf "joinu-im-node/proto"
	"net"
	"net/http"
	"tursom-im/attr"
	"tursom-im/context"
)

type WebSocketHandler struct {
	globalContext context.GlobalContext
}

func NewWebSocketHandler(globalContext context.GlobalContext) *WebSocketHandler {
	return &WebSocketHandler{
		globalContext: globalContext,
	}
}

func (c *WebSocketHandler) UpgradeToWebSocket(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		fmt.Println(err)
		return
	}
	go c.Handle(conn)
}

func (c *WebSocketHandler) Handle(conn net.Conn) {
	defer func() {
		err2 := conn.Close()
		fmt.Println(err2)
	}()

	attachmentConn := attr.NewSimpleAttachmentConn(&conn)

	for {
		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil {
			fmt.Println(err)
			return
		}

		switch op {
		case ws.OpBinary:
			imMsg := com_joinu_im_protobuf.ImMsg{}
			err := proto.Unmarshal(msg, &imMsg)
			if err != nil {
				fmt.Println(err)
				continue
			}
			c.handleBinaryMsg(attachmentConn, &imMsg)
		case ws.OpText:
		case ws.OpPing:
		}
	}
}

func (c *WebSocketHandler) handleBinaryMsg(conn *attr.AttachmentConn, msg *com_joinu_im_protobuf.ImMsg) {
	fmt.Println(msg)
	imMsg := com_joinu_im_protobuf.ImMsg{
		CmdId: msg.CmdId,
	}

	switch x := msg.GetContent().(type) {
	case *com_joinu_im_protobuf.ImMsg_C2CMsgRequest:
	case *com_joinu_im_protobuf.ImMsg_GroupMsgRequest:
	case *com_joinu_im_protobuf.ImMsg_LoginRequest:
		fmt.Println(x.LoginRequest)
		loginResult := c.handleBinaryLogin(conn, msg)
		imMsg.Content = loginResult
	case *com_joinu_im_protobuf.ImMsg_OfflineMsgRequest:
	}
	bytes, err := proto.Marshal(&imMsg)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsutil.WriteServerBinary(conn, bytes)
}

func (c *WebSocketHandler) handleBinaryLogin(conn *attr.AttachmentConn, msg *com_joinu_im_protobuf.ImMsg) (loginResult *com_joinu_im_protobuf.ImMsg_LoginResult) {
	loginResult = &com_joinu_im_protobuf.ImMsg_LoginResult{
		LoginResult: &com_joinu_im_protobuf.LoginResult{},
	}

	token, err := c.globalContext.TokenContext().Parse(msg.GetLoginRequest().Token)
	if err != nil {
		fmt.Println(err)
		return
	}
	//TODO 从redis校验token
	userIdAttr := conn.Get(c.globalContext.AttrContext().UserIdAttrKey())
	userTokenAttr := conn.Get(c.globalContext.AttrContext().UserTokenAttrKey())
	err = userIdAttr.Set(token.Uid)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = userTokenAttr.Set(token)
	if err != nil {
		fmt.Println(err)
		return
	}
	loginResult.LoginResult.ImUserId = token.Uid
	loginResult.LoginResult.Success = true

	return
}
