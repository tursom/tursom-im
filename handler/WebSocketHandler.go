package handler

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/protobuf/proto"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"tursom-im/context"
	"tursom-im/im_conn"
	"tursom-im/proto"
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

	attachmentConn := im_conn.NewSimpleAttachmentConn(&conn)

	for {
		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil {
			fmt.Println(err)
			return
		}

		switch op {
		case ws.OpBinary:
			imMsg := tursom_im_protobuf.ImMsg{}
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

func (c *WebSocketHandler) handleBinaryMsg(conn *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg) {
	fmt.Println(msg)
	imMsg := tursom_im_protobuf.ImMsg{}

	switch msg.GetContent().(type) {
	case *tursom_im_protobuf.ImMsg_SendMsgRequest:
	case *tursom_im_protobuf.ImMsg_LoginRequest:
		loginResult := c.handleBinaryLogin(conn, msg)
		imMsg.Content = loginResult
	}
	bytes, err := proto.Marshal(&imMsg)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsutil.WriteServerBinary(conn, bytes)
}

func (c *WebSocketHandler) handleBinaryLogin(conn *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg) (loginResult *tursom_im_protobuf.ImMsg_LoginResult) {
	loginResult = &tursom_im_protobuf.ImMsg_LoginResult{
		LoginResult: &tursom_im_protobuf.LoginResult{},
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
