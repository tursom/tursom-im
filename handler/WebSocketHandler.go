package handler

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/protobuf/proto"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"os"
	"runtime"
	"tursom-im/context"
	"tursom-im/im_conn"
	"tursom-im/tursom_im_protobuf"
)

type WebSocketHandler struct {
	globalContext *context.GlobalContext
}

func NewWebSocketHandler(globalContext *context.GlobalContext) *WebSocketHandler {
	return &WebSocketHandler{
		globalContext: globalContext,
	}
}

func (c *WebSocketHandler) InitWebHandler(basePath string, router *httprouter.Router) {
	router.GET(basePath+"/ws", c.UpgradeToWebSocket)
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
	defer conn.Close()

	attachmentConn := im_conn.NewSimpleAttachmentConn(&conn)

	for {
		err := c.loop(attachmentConn)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func (c *WebSocketHandler) loop(conn *im_conn.AttachmentConn) (err error) {
	//goland:noinspection GoUnhandledErrorResult
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintln(os.Stderr, "an panic caused on handle WebSocket message")
			fmt.Fprintln(os.Stderr, err)
			for i := 1; ; i++ {
				pc, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				fmt.Fprintln(os.Stderr, pc, file, line)
			}
		}
	}()

	msg, op, err := wsutil.ReadClientData(conn)
	if err != nil {
		return err
	}

	if !op.IsData() {
		return nil
	}

	switch op {
	case ws.OpBinary:
		imMsg := tursom_im_protobuf.ImMsg{}
		err = proto.Unmarshal(msg, &imMsg)
		if err != nil {
			return err
		}
		c.handleBinaryMsg(conn, &imMsg)
	case ws.OpText:
		panic("could not handle text message")
	default:
		panic("could not handle unknown message")
	}

	return nil
}

func (c *WebSocketHandler) handleBinaryMsg(conn *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg) {
	fmt.Println(msg)
	imMsg := tursom_im_protobuf.ImMsg{}
	closeConnection := false
	defer func() {
		if closeConnection {
			err := conn.Close()
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	if msg.SelfMsg {
		c.handleSelfMsg(conn, msg)
		return
	}

	switch msg.GetContent().(type) {
	case *tursom_im_protobuf.ImMsg_SendMsgRequest:
		imMsg.Content, imMsg.MsgId = c.handleSendChatMsg(conn, msg)
	case *tursom_im_protobuf.ImMsg_LoginRequest:
		loginResult := c.handleBinaryLogin(conn, msg)
		imMsg.Content = loginResult
		closeConnection = !loginResult.LoginResult.Success
	}
	bytes, err := proto.Marshal(&imMsg)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = wsutil.WriteServerBinary(conn, bytes)
	if err != nil {
		fmt.Println(err)
	}
}

func (c *WebSocketHandler) handleSelfMsg(conn *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg) {
	sender := conn.Get(c.globalContext.AttrContext().UserIdAttrKey()).Get().(string)
	currentConn := c.globalContext.UserConnContext().GetUserConn(sender)
	currentConn.WriteChatMsg(msg, func(c *im_conn.AttachmentConn) bool {
		return conn != c
	})
}

func (c *WebSocketHandler) handleSendChatMsg(conn *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg) (response *tursom_im_protobuf.ImMsg_SendMsgResponse, msgId string) {
	response = &tursom_im_protobuf.ImMsg_SendMsgResponse{SendMsgResponse: &tursom_im_protobuf.SendMsgResponse{}}
	msgId = c.globalContext.MsgIdContext().NewMsgIdStr()
	sendMsgRequest := msg.GetSendMsgRequest()

	sender := conn.Get(c.globalContext.AttrContext().UserIdAttrKey()).Get().(string)

	response.SendMsgResponse.ReqId = sendMsgRequest.ReqId

	receiver := sendMsgRequest.Receiver
	receiverConn := c.globalContext.UserConnContext().GetUserConn(receiver)
	currentConn := c.globalContext.UserConnContext().GetUserConn(sender)
	if receiverConn == nil || currentConn == nil {
		response.SendMsgResponse.FailMsg = "user \"" + receiver + "\" not login"
		response.SendMsgResponse.FailType = tursom_im_protobuf.FailType_TARGET_NOT_LOGIN
		return
	}

	imMsg := &tursom_im_protobuf.ImMsg{
		MsgId: msgId,
		Content: &tursom_im_protobuf.ImMsg_ChatMsg{ChatMsg: &tursom_im_protobuf.ChatMsg{
			Receiver: receiver,
			Sender:   sender,
			Content:  sendMsgRequest.Content,
		}},
	}
	receiverConn.WriteChatMsg(imMsg, nil)
	currentConn.WriteChatMsg(imMsg, func(c *im_conn.AttachmentConn) bool {
		return conn != c
	})

	return
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

	c.globalContext.UserConnContext().TouchUserConn(token.Uid).Add(conn)

	loginResult.LoginResult.ImUserId = token.Uid
	loginResult.LoginResult.Success = true

	return
}
