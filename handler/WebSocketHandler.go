package handler

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"sync/atomic"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/protobuf/proto"
	"github.com/julienschmidt/httprouter"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util"

	"github.com/tursom-im/context"
	"github.com/tursom-im/exception"
	"github.com/tursom-im/im_conn"
	"github.com/tursom-im/tursom_im_protobuf"
	"github.com/tursom-im/utils"
)

var (
	imHandlerFactories []func(ctx *context.GlobalContext) ImMsgHandler
	imHandlerContext   = util.NewContext()
	ResponseCtxKey     = util.AllocateContextKeyWithDefault[*tursom_im_protobuf.ImMsg](imHandlerContext, func() *tursom_im_protobuf.ImMsg {
		return &tursom_im_protobuf.ImMsg{}
	})
	CloseConnectionCtxKey = util.AllocateContextKey[bool](imHandlerContext)
)

type (
	// ImMsgHandler shows an object that can handle im msg
	// default im handlers on package handler/msg, you need to run msg.Init() to initial imHandlerFactories
	ImMsgHandler interface {
		lang.Object
		HandleMsg(
			conn *im_conn.AttachmentConn,
			msg *tursom_im_protobuf.ImMsg,
			ctx util.ContextMap,
		) (ok bool)
	}

	WebSocketHandler struct {
		lang.BaseObject
		globalContext     *context.GlobalContext
		writeChannelList  []chan *im_conn.ConnWriteMsg
		writeChannelIndex uint32
		handlers          []ImMsgHandler
	}
)

func RegisterImHandlerFactory(handlerFactory func(ctx *context.GlobalContext) ImMsgHandler) {
	imHandlerFactories = append(imHandlerFactories, handlerFactory)
}

func GetImMsgHandlers(ctx *context.GlobalContext) []ImMsgHandler {
	handlers := make([]ImMsgHandler, len(imHandlerFactories))
	for i, factory := range imHandlerFactories {
		handlers[i] = factory(ctx)
	}
	return handlers
}

func NewWebSocketHandler(globalContext *context.GlobalContext) *WebSocketHandler {
	return &WebSocketHandler{
		globalContext:     globalContext,
		writeChannelList:  nil,
		writeChannelIndex: 0,
		handlers:          GetImMsgHandlers(globalContext),
	}
}

func (h *WebSocketHandler) InitWebHandler(router Router) {
	exceptions.CheckNil(h)

	if h.writeChannelList == nil {
		writeChannelCount := int(math.Max(16, float64(runtime.NumCPU()*2)))
		for i := 0; i < writeChannelCount; i++ {
			writeChannel := make(chan *im_conn.ConnWriteMsg, 128)
			go im_conn.HandleWrite(writeChannel)
			h.writeChannelList = append(h.writeChannelList, writeChannel)
		}
	}

	router.GET("/ws", h.UpgradeToWebSocket)
}

func (h *WebSocketHandler) UpgradeToWebSocket(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	exceptions.CheckNil(h)

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		exceptions.Package(err).PrintStackTrace()
		return
	}
	go h.Handle(conn)
}

func (h *WebSocketHandler) Handle(conn net.Conn) {
	exceptions.CheckNil(h)

	writeChannelIndex := atomic.AddUint32(&h.writeChannelIndex, 1) % uint32(len(h.writeChannelList))
	attachmentConn := im_conn.NewSimpleAttachmentConn(conn, h.writeChannelList[writeChannelIndex])
	//goland:noinspection GoUnhandledErrorResult
	defer attachmentConn.Close()

	watchDog := utils.NewWatchDog(60, func() {
		_ = attachmentConn.Close()
	})
	if watchDog == nil {
		exceptions.PackageAny("watch dog register failed").PrintStackTrace()
		return
	}
	defer watchDog.Kill()

	for {
		_, err := exceptions.Try(func() (any, exceptions.Exception) {
			msg, op, err := wsutil.ReadClientData(conn)
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
					_, err := exceptions.Try(func() (any, exceptions.Exception) {
						h.handleBinaryMsg(attachmentConn, imMsg)
						return nil, nil
					}, func(panic any) (any, exceptions.Exception) {
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
		}, func(panic any) (any, exceptions.Exception) {
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

func (h *WebSocketHandler) handleBinaryMsg(conn *im_conn.AttachmentConn, request *tursom_im_protobuf.ImMsg) {
	exceptions.CheckNil(h)

	sender, login := h.globalContext.AttrContext().UserIdAttrKey().Get(conn).TryGet()
	if login {
		fmt.Println("request:", sender, ":", request)
	} else {
		fmt.Println("request:", request)
	}
	ctx := imHandlerContext.NewMap()
	defer func() {
		if CloseConnectionCtxKey.Get(ctx) {
			_ = conn.Close()
		}
	}()

	if request.SelfMsg {
		h.handleSelfMsg(conn, request)
		return
	}

	for _, handler := range h.handlers {
		if handler.HandleMsg(conn, request, ctx) {
			break
		}
	}

	response := ResponseCtxKey.Get(ctx)
	if login {
		fmt.Println("response:", sender, ":", response)
	} else {
		fmt.Println("response:", response)
	}
	bytes, err := proto.Marshal(response)
	if err != nil {
		exceptions.Print(err)
		return
	}
	conn.WriteData(bytes)
}

func (h *WebSocketHandler) handleSelfMsg(conn *im_conn.AttachmentConn, msg *tursom_im_protobuf.ImMsg) {
	exceptions.CheckNil(h)

	sender := h.globalContext.AttrContext().UserIdAttrKey().Get(conn).Get().AsString()
	currentConn := h.globalContext.UserConnContext().GetUserConn(sender)
	_ = currentConn.WriteChatMsg(msg, func(c *im_conn.AttachmentConn) bool {
		return conn != c
	})
}
