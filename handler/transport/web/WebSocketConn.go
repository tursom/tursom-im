package web

import (
	"net"
	"sync"
	"time"

	"github.com/gobwas/ws/wsutil"
	"github.com/golang/protobuf/proto"
	"github.com/tursom/GoCollections/collections"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"

	cc "github.com/tursom/GoCollections/concurrent/collections"

	"github.com/tursom-im/conn"
	"github.com/tursom-im/exception"
	m "github.com/tursom-im/proto/msg"
)

//goland:noinspection GoNameStartsWithPackageName
type (
	EventListener struct {
		lang.BaseObject
		listener func(event conn.Event)
		node     collections.QueueNode[*EventListener]
	}

	ConnWriteMsg struct {
		lang.BaseObject
		Conn       *WebSocketConn
		Data       []byte
		ErrHandler func(err exceptions.Exception)
	}

	Attachment[T any] struct {
		lang.BaseObject
		key        *conn.AttachmentKey[T]
		attachment *sync.Map
	}

	WebSocketConn struct {
		lang.BaseObject
		conn              net.Conn
		writeChannel      chan *ConnWriteMsg
		attachment        sync.Map
		eventListenerList cc.ConcurrentLinkedQueue[*EventListener]
	}
)

func (c *WebSocketConn) SendMsg(msg *m.ImMsg) {
	bytes, err := proto.Marshal(msg)
	if err != nil {
		panic(exceptions.Package(err))
	}

	c.WriteData(bytes)
}

func (c *WebSocketConn) Attr(a *conn.AttachmentKey[any]) conn.Attachment[any] {
	return &Attachment[any]{
		key:        a,
		attachment: &c.attachment,
	}
}

func HandleWrite(writeChannel <-chan *ConnWriteMsg) {
	for writeMsg := range writeChannel {
		if _, err := exceptions.Try[any](func() (any, exceptions.Exception) {
			if writeErr := wsutil.WriteServerBinary(writeMsg.Conn, writeMsg.Data); writeErr != nil {
				err := exceptions.Package(writeErr)
				if writeMsg.ErrHandler != nil {
					writeMsg.ErrHandler(err)
				} else {
					exception.Log("failed to write msg", err)
				}
			}
			return nil, nil
		}, func(panic any) (ret any, err exceptions.Exception) {
			return nil, exceptions.PackageAny(panic)
		}); err != nil {
			exception.Log("failed to write msg", err)
		}
	}
}

func (l *EventListener) Remove() exceptions.Exception {
	if l == nil {
		return nil
	}
	return l.node.Remove()
}

func (c *WebSocketConn) WriteChannel() chan<- *ConnWriteMsg {
	return c.writeChannel
}

func (c *WebSocketConn) WriteData(data []byte) {
	c.writeChannel <- &ConnWriteMsg{Conn: c, Data: data}
}

func NewSimpleAttachmentConn(conn net.Conn, writeChannel chan *ConnWriteMsg) *WebSocketConn {
	return &WebSocketConn{
		conn:         conn,
		writeChannel: writeChannel,
	}
}

func (c *WebSocketConn) notify(event conn.Event) {
	if c == nil {
		panic(exceptions.NewNPE("WebSocketConn is null", nil))
	}
	err := collections.Loop[*EventListener](&c.eventListenerList, func(element *EventListener) exceptions.Exception {
		_, _ = exceptions.Try(func() (ret any, err exceptions.Exception) {
			element.listener(event)
			return
		}, func(panic any) (ret any, err exceptions.Exception) {
			exceptions.NewRuntimeException("", exceptions.DefaultExceptionConfig().SetCause(panic)).PrintStackTrace()
			return
		})
		return nil
	})
	exceptions.Print(err)
}

func (c *WebSocketConn) AddEventListener(f func(event conn.Event)) conn.EventListener {
	if c == nil {
		panic(exceptions.NewNPE("WebSocketConn is null", nil))
	}
	if f == nil {
		return nil
	}
	listener := &EventListener{listener: f}
	listener.node = exceptions.Exec1r1(c.eventListenerList.OfferAndGetNode, listener)
	return listener
}

func (c *WebSocketConn) Read(b []byte) (n int, err error) {
	if c == nil {
		return 0, exceptions.NewNPE("WebSocketConn is null", nil)
	}
	read, err := c.conn.Read(b)
	return read, exceptions.Package(err)
}

func (c *WebSocketConn) Write(b []byte) (n int, err error) {
	if c == nil {
		return 0, exceptions.NewNPE("WebSocketConn is null", nil)
	}
	write, err := c.conn.Write(b)
	return write, exceptions.Package(err)
}

func (c *WebSocketConn) Close() error {
	if c == nil {
		panic(exceptions.NewNPE("WebSocketConn is null", nil))
	}
	go c.notify(conn.NewEventClosed(c))
	return exceptions.Package(c.conn.Close())
}

func (c *WebSocketConn) LocalAddr() net.Addr {
	if c == nil {
		panic(exceptions.NewNPE("WebSocketConn is null", nil))
	}
	return c.conn.LocalAddr()
}

func (c *WebSocketConn) RemoteAddr() net.Addr {
	if c == nil {
		panic(exceptions.NewNPE("WebSocketConn is null", nil))
	}
	return c.conn.RemoteAddr()
}

func (c *WebSocketConn) SetDeadline(t time.Time) error {
	if c == nil {
		panic(exceptions.NewNPE("WebSocketConn is null", nil))
	}
	return exceptions.Package(c.conn.SetDeadline(t))
}

func (c *WebSocketConn) SetReadDeadline(t time.Time) error {
	if c == nil {
		panic(exceptions.NewNPE("WebSocketConn is null", nil))
	}
	return exceptions.Package(c.conn.SetReadDeadline(t))
}

func (c *WebSocketConn) SetWriteDeadline(t time.Time) error {
	if c == nil {
		panic(exceptions.NewNPE("WebSocketConn is null", nil))
	}
	return exceptions.Package(c.conn.SetWriteDeadline(t))
}

func (a *Attachment[T]) Get() T {
	if a == nil {
		panic(exceptions.NewNPE("WebSocketConn is null", nil))
	}
	load, _ := a.attachment.Load(a.key.Id())
	value, ok := load.(T)
	if !ok {
		panic(exceptions.NewTypeCastExceptionByType[T](load, nil))
	}
	return value
}

func (a *Attachment[T]) TryGet() (T, bool) {
	if a == nil {
		panic(exceptions.NewNPE("WebSocketConn is null", nil))
	}
	load, _ := a.attachment.Load(a.key.Id())
	value, ok := load.(T)
	return value, ok
}

func (a *Attachment[T]) Set(value T) {
	if a == nil {
		panic(exceptions.NewNPE("WebSocketConn is null", nil))
	}
	a.attachment.Store(a.key.Id(), value)
}
