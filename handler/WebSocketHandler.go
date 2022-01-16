package handler

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/golang/protobuf/proto"
	"github.com/julienschmidt/httprouter"
	"github.com/tursom/GoCollections/exceptions"
	"net"
	"net/http"
	"os"
	"reflect"
	"time"
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
	if c == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", true))
	}

	router.GET(basePath+"/ws", c.UpgradeToWebSocket)
}

func (c *WebSocketHandler) UpgradeToWebSocket(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if c == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", true))
	}

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		exceptions.Package(err).PrintStackTrace()
		return
	}
	go c.Handle(conn)
}

func (c *WebSocketHandler) Handle(conn net.Conn) {
	if c == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", true))
	}

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

	go attachmentConn.LoopRead()
	for {
		_, err := exceptions.Try(func() (interface{}, exceptions.Exception) {
			err := attachmentConn.HandleWrite()
			if err != nil {
				return nil, exceptions.Package(err)
			}
			msg, op, err := attachmentConn.TryRead()
			if msg == nil && err == nil {
				time.Sleep(500 * time.Millisecond)
				return nil, nil
			}

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
						exceptions.Print(err)
						if !utils.IsClosedError(err) {
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
			exceptions.Print(err)
			if utils.IsClosedError(err) {
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
	if c == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", true))
	}

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
	conn.WriteChannel() <- bytes
}

func (c *WebSocketHandler) handleSelfMsg(conn *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg) {
	if c == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", true))
	}

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
	if c == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", true))
	}

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
	if c == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", true))
	}

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
	if c == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", true))
	}

	sendBroadcastRequest := msg.GetSendBroadcastRequest()
	response := &tursom_im_protobuf.SendBroadcastResponse{
		ReqId: sendBroadcastRequest.ReqId,
	}

	broadcast := &tursom_im_protobuf.ImMsg{
		MsgId: c.globalContext.MsgIdContext().NewMsgIdStr(),
		Content: &tursom_im_protobuf.ImMsg_Broadcast{Broadcast: &tursom_im_protobuf.Broadcast{
			Sender:  conn.Get(c.globalContext.AttrContext().UserIdAttrKey()).Get().(string),
			ReqId:   sendBroadcastRequest.ReqId,
			Channel: sendBroadcastRequest.Channel,
			Content: sendBroadcastRequest.Content,
		}},
	}
	bytes, err := proto.Marshal(broadcast)
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
	if c == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", true))
	}

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
	_ = currentConn.Aggregation(receiverConn).WriteChatMsg(imMsg, func(c *im_conn.AttachmentConn) bool {
		return conn != c
	})

	return
}

func (c *WebSocketHandler) handleBinaryLogin(
	conn *im_conn.AttachmentConn,
	msg *tursom_im_protobuf.ImMsg,
) (loginResult *tursom_im_protobuf.ImMsg_LoginResult) {
	if c == nil {
		panic(exceptions.NewNPE("WebSocketHandler is null", true))
	}
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

	uid := token.Uid
	if msg.GetLoginRequest().TempId {
		uid = uid + "-" + c.globalContext.MsgIdContext().NewMsgIdStr()
	}

	err = userIdAttr.Set(uid)
	if err != nil {
		exceptions.Print(err)
		return
	}
	err = userTokenAttr.Set(token)
	if err != nil {
		exceptions.Print(err)
		return
	}

	connGroup := c.globalContext.UserConnContext().TouchUserConn(uid)
	connGroup.Add(conn)
	conn.AddEventListener(func(event im_conn.ConnEvent) {
		if event.EventId().IsConnClosed() {
			if connGroup.Size() == 0 {
				c.globalContext.UserConnContext().RemoveUserConn(uid)
			}
		}
	})

	loginResult.LoginResult.ImUserId = uid
	loginResult.LoginResult.Success = true

	return
}
