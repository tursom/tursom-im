package handler

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/protobuf/proto"
	"github.com/julienschmidt/httprouter"
	"github.com/tursom/GoCollections/exceptions"
	"net"
	"net/http"
	"os"
	"reflect"
	"tursom-im/context"
	"tursom-im/exception"
	"tursom-im/im_conn"
	"tursom-im/tursom_im_protobuf"
	"tursom-im/utils"
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
		exceptions.Package(err).PrintStackTrace()
		return
	}
	go c.Handle(conn)
}

func (c *WebSocketHandler) Handle(conn net.Conn) {
	attachmentConn := im_conn.NewSimpleAttachmentConn(conn)
	//goland:noinspection GoUnhandledErrorResult
	defer attachmentConn.Close()

	watchDog := utils.NewWatchDog(60, func() bool {
		_ = attachmentConn.Close()
		return true
	})
	if watchDog == nil {
		exceptions.PackageAny("watch dog register failed").PrintStackTrace()
		return
	}

	for {
		_, err := exceptions.Try(func() (interface{}, exceptions.Exception) {
			msg, op, err := wsutil.ReadClientData(attachmentConn)
			if err != nil {
				return nil, exceptions.Package(err)
			}
			watchDog.Feed()

			if !op.IsData() {
				return nil, nil
			}

			switch op {
			case ws.OpBinary:
				imMsg := &tursom_im_protobuf.ImMsg{}
				err = proto.Unmarshal(msg, imMsg)
				if err != nil {
					return nil, exceptions.Package(err)
				}
				go func() {
					_, err := exceptions.Try(func() (interface{}, exceptions.Exception) {
						c.handleBinaryMsg(attachmentConn, imMsg)
						return nil, nil
					}, func(panic interface{}) (interface{}, exceptions.Exception) {
						return nil, exceptions.PackagePanic(panic, "an panic caused on handle WebSocket message:")
					})
					if err != nil {
						if !utils.IsClosedError(err) {
							exceptions.Print(err)
							exceptions.Print(conn.Close())
						}
					}
				}()
			case ws.OpText:
				exception.NewUnsupportedException("could not handle text message").PrintStackTrace()
			default:
				exception.NewUnsupportedException("could not handle unknown message").PrintStackTrace()
			}
			return nil, nil
		}, func(panic interface{}) (interface{}, exceptions.Exception) {
			return nil, exceptions.PackagePanic(panic, "an panic caused on handle WebSocket message:")
		})
		if err != nil {
			if !utils.IsClosedError(err) {
				exceptions.Print(err)
				unpack := exceptions.UnpackException(err)
				if unpack == nil {
					_, _ = fmt.Fprintln(os.Stderr, "error type:", reflect.TypeOf(unpack))
				}
			}
			return
		}
	}
}

func (c *WebSocketHandler) handleBinaryMsg(conn *im_conn.AttachmentConn, request *tursom_im_protobuf.ImMsg) {
	sender := conn.Get(c.globalContext.AttrContext().UserIdAttrKey()).Get()
	fmt.Println("request:", sender, ":", request)
	response := tursom_im_protobuf.ImMsg{}
	closeConnection := false
	defer func() {
		if closeConnection {
			_ = conn.Close()
		}
	}()

	if request.SelfMsg {
		c.handleSelfMsg(conn, request)
		return
	}

	switch request.GetContent().(type) {
	case *tursom_im_protobuf.ImMsg_SendMsgRequest:
		response.Content, response.MsgId = c.handleSendChatMsg(conn, request)
	case *tursom_im_protobuf.ImMsg_LoginRequest:
		loginResult := c.handleBinaryLogin(conn, request)
		response.Content = loginResult
		closeConnection = !loginResult.LoginResult.Success
	case *tursom_im_protobuf.ImMsg_HeartBeat:
		response.Content = request.Content
	case *tursom_im_protobuf.ImMsg_AllocateNodeRequest:
		response.Content = c.handleAllocateNode(conn, request)
	case *tursom_im_protobuf.ImMsg_SendBroadcastRequest:
		response.Content = c.handleSendBroadcast(conn, request)
	case *tursom_im_protobuf.ImMsg_ListenBroadcastRequest:
		response.Content = c.handleListenBroadcast(conn, request)
	}

	fmt.Println("response:", sender, ":", &response)
	bytes, err := proto.Marshal(&response)
	if err != nil {
		exceptions.Print(err)
		return
	}
	err = wsutil.WriteServerBinary(conn, bytes)
	if err != nil {
		exceptions.Print(err)
	}
}

func (c *WebSocketHandler) handleSelfMsg(conn *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg) {
	sender := conn.Get(c.globalContext.AttrContext().UserIdAttrKey()).Get().(string)
	currentConn := c.globalContext.UserConnContext().GetUserConn(sender)
	_ = currentConn.WriteChatMsg(msg, func(c *im_conn.AttachmentConn) bool {
		return conn != c
	})
}

func (c *WebSocketHandler) handleAllocateNode(
	conn *im_conn.AttachmentConn,
	msg *tursom_im_protobuf.ImMsg,
) *tursom_im_protobuf.ImMsg_AllocateNodeResponse {
	allocateNodeResponse := &tursom_im_protobuf.AllocateNodeResponse{
		ReqId: msg.GetAllocateNodeRequest().ReqId,
	}

	allocateNodeResponse.Node = c.globalContext.ConnNodeContext().Allocate(conn)

	return &tursom_im_protobuf.ImMsg_AllocateNodeResponse{
		AllocateNodeResponse: allocateNodeResponse,
	}
}

func (c *WebSocketHandler) handleListenBroadcast(
	conn *im_conn.AttachmentConn,
	msg *tursom_im_protobuf.ImMsg,
) *tursom_im_protobuf.ImMsg_ListenBroadcastResponse {
	listenBroadcastRequest := msg.GetListenBroadcastRequest()
	response := &tursom_im_protobuf.ListenBroadcastResponse{
		ReqId: listenBroadcastRequest.ReqId,
	}

	var err error = nil
	if listenBroadcastRequest.CancelListen {
		err = c.globalContext.BroadcastContext().CancelListen(listenBroadcastRequest.Channel, conn)
	} else {
		err = c.globalContext.BroadcastContext().Listen(listenBroadcastRequest.Channel, conn)
	}
	if err != nil {
		exceptions.Print(err)
	} else {
		response.Success = true
	}
	return &tursom_im_protobuf.ImMsg_ListenBroadcastResponse{
		ListenBroadcastResponse: response,
	}
}

func (c *WebSocketHandler) handleSendBroadcast(
	conn *im_conn.AttachmentConn,
	msg *tursom_im_protobuf.ImMsg,
) *tursom_im_protobuf.ImMsg_SendBroadcastResponse {
	sendBroadcastRequest := msg.GetSendBroadcastRequest()
	response := &tursom_im_protobuf.SendBroadcastResponse{
		ReqId: sendBroadcastRequest.ReqId,
	}

	imMsg := &tursom_im_protobuf.ImMsg{
		MsgId: c.globalContext.MsgIdContext().NewMsgIdStr(),
		Content: &tursom_im_protobuf.ImMsg_Broadcast{Broadcast: &tursom_im_protobuf.Broadcast{
			Sender:  conn.Get(c.globalContext.AttrContext().UserIdAttrKey()).Get().(string),
			ReqId:   sendBroadcastRequest.ReqId,
			Channel: sendBroadcastRequest.Channel,
			Content: sendBroadcastRequest.Content,
		}},
	}
	bytes, err := proto.Marshal(imMsg)
	if err != nil {
		exceptions.Print(err)
	} else {
		response.ReceiverCount = c.globalContext.BroadcastContext().Send(
			sendBroadcastRequest.Channel,
			bytes,
			//nil,
			func(c *im_conn.AttachmentConn) bool {
				return c != conn
			},
		)
	}
	return &tursom_im_protobuf.ImMsg_SendBroadcastResponse{
		SendBroadcastResponse: response,
	}
}

func (c *WebSocketHandler) handleSendChatMsg(
	conn *im_conn.AttachmentConn,
	msg *tursom_im_protobuf.ImMsg,
) (
	response *tursom_im_protobuf.ImMsg_SendMsgResponse,
	msgId string,
) {
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

	response.SendMsgResponse.Success = true
	imMsg := &tursom_im_protobuf.ImMsg{
		MsgId: msgId,
		Content: &tursom_im_protobuf.ImMsg_ChatMsg{ChatMsg: &tursom_im_protobuf.ChatMsg{
			Receiver: receiver,
			Sender:   sender,
			Content:  sendMsgRequest.Content,
		}},
	}
	_ = receiverConn.WriteChatMsg(imMsg, nil)
	_ = currentConn.WriteChatMsg(imMsg, func(c *im_conn.AttachmentConn) bool {
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
		exceptions.Print(err)
		return
	}

	userIdAttr := conn.Get(c.globalContext.AttrContext().UserIdAttrKey())
	userTokenAttr := conn.Get(c.globalContext.AttrContext().UserTokenAttrKey())
	err = userIdAttr.Set(token.Uid)
	if err != nil {
		exceptions.Print(err)
		return
	}
	err = userTokenAttr.Set(token)
	if err != nil {
		exceptions.Print(err)
		return
	}

	c.globalContext.UserConnContext().TouchUserConn(token.Uid).Add(conn)

	loginResult.LoginResult.ImUserId = token.Uid
	loginResult.LoginResult.Success = true

	return
}
