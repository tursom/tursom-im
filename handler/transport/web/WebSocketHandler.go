package web

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
	log "github.com/sirupsen/logrus"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"

	"github.com/tursom/tursom-im/context"
	"github.com/tursom/tursom-im/exception"
	"github.com/tursom/tursom-im/handler"
	m "github.com/tursom/tursom-im/proto/msg"
	"github.com/tursom/tursom-im/proto/msys"
	"github.com/tursom/tursom-im/utils"
)

type (
	// Handler WebSocket handler
	Handler struct {
		lang.BaseObject
		globalContext     *context.GlobalContext
		writeChannelList  []chan *ConnWriteMsg
		writeChannelIndex uint32
		handlers          []handler.IMMsgHandler
	}
)

func NewWebSocketHandler(globalContext *context.GlobalContext) *Handler {
	return &Handler{
		globalContext:     globalContext,
		writeChannelList:  nil,
		writeChannelIndex: 0,
		handlers:          handler.MsgHandlers(globalContext),
	}
}

func (h *Handler) InitWebHandler(router Router) {
	exceptions.CheckNil(h)

	if h.writeChannelList == nil {
		writeChannelCount := int(math.Max(16, float64(runtime.NumCPU()*2)))
		for i := 0; i < writeChannelCount; i++ {
			writeChannel := make(chan *ConnWriteMsg, 128)
			go HandleWrite(writeChannel)
			h.writeChannelList = append(h.writeChannelList, writeChannel)
		}
	}

	router.GET("/ws", h.clientUpgrade)
}

func (h *Handler) clientUpgrade(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	exceptions.CheckNil(h)

	c, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		exceptions.Package(err).PrintStackTrace()
		return
	}
	go h.Handle(c)
}

func (h *Handler) Handle(conn net.Conn) {
	exceptions.CheckNil(h)

	writeChannelIndex := atomic.AddUint32(&h.writeChannelIndex, 1) % uint32(len(h.writeChannelList))
	attachmentConn := NewSimpleAttachmentConn(conn, h.writeChannelList[writeChannelIndex])
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
		if _, err := exceptions.Try(func() (any, exceptions.Exception) {
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
				imMsg := &m.ImMsg{}
				if err = proto.Unmarshal(msg, imMsg); err != nil {
					return nil, exceptions.Package(err)
				}
				go func() {
					if _, err := exceptions.Try(func() (any, exceptions.Exception) {
						h.handleBinaryMsg(attachmentConn, imMsg)
						return nil, nil
					}, func(panic any) (any, exceptions.Exception) {
						return nil, exceptions.PackagePanic(panic, "an panic caused on handle WebSocket message:")
					}); err != nil {
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
		}); err != nil {
			exceptions.Print(err)
			if !utils.IsClosedError(err) {
				return
			}

			if unpack := exceptions.UnpackException(err); unpack == nil {
				_, _ = fmt.Fprintln(os.Stderr, "error type:", reflect.TypeOf(unpack))
			}
		}
	}
}

func (h *Handler) handleBinaryMsg(conn *WebSocketConn, request *m.ImMsg) {
	exceptions.CheckNil(h)

	sender, login := h.globalContext.Attr().UserId(conn).TryGet()
	logRequest(request, sender, login)

	ctx := handler.NewImMsgContext()
	defer func() {
		if handler.CloseConnectionCtxKey.Get(ctx) {
			_ = conn.Close()
		}
	}()

	if request.SelfMsg {
		h.handleSelfMsg(conn, request)
		return
	}

	for _, msgHandler := range h.handlers {
		if msgHandler.HandleMsg(conn, request, ctx) {
			break
		}
	}

	response := handler.ResponseCtxKey.Get(ctx)
	logResponse(login, sender, response)

	bytes, err := proto.Marshal(response)
	if err != nil {
		exception.Log("marshal response failed", err)
		return
	}
	conn.WriteData(bytes)
}

func logRequest(request *m.ImMsg, sender lang.String, login bool) {
	if login {
		// if already login
		log.WithField("request", request).
			WithField("sender", sender).
			Info("request")
	} else {
		log.WithField("request", request).
			Info("request")
	}
}

func logResponse(login bool, sender lang.String, response *m.ImMsg) {
	if login {
		log.WithField("response", response).
			WithField("sender", sender).
			Info("response")
	} else {
		log.WithField("response", response).
			Info("response")
	}
}

func (h *Handler) handleSelfMsg(c *WebSocketConn, msg *m.ImMsg) {
	exceptions.CheckNil(h)

	sender := h.globalContext.Attr().UserId(c).Get().String()
	h.globalContext.Broadcast().Send(uint32(msys.Channel_USER), sender, msg)
}
