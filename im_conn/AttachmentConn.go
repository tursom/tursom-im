package im_conn

import (
	"github.com/gobwas/ws/wsutil"
	"github.com/tursom-im/exception"
	"github.com/tursom/GoCollections/collections"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var attachmentKeyId = int32(0)

type (
	AttachmentKey[T any] struct {
		lang.BaseObject
		name string
		id   int32
	}

	ConnWriteMsg struct {
		lang.BaseObject
		Conn       *AttachmentConn
		Data       []byte
		ErrHandler func(err exceptions.Exception)
	}

	Attachment[T any] struct {
		lang.BaseObject
		key        *AttachmentKey[T]
		attachment *sync.Map
	}

	EventListener struct {
		lang.BaseObject
		listener func(event ConnEvent)
		node     collections.ConcurrentLinkedQueueNode[*EventListener]
	}

	AttachmentConn struct {
		lang.BaseObject
		attachment        *sync.Map
		conn              net.Conn
		eventListenerList collections.ConcurrentLinkedQueue[*EventListener]
		writeChannel      chan ConnWriteMsg
	}
)

func HandleWrite(writeChannel <-chan ConnWriteMsg) {
	for true {
		_, err := exceptions.Try(func() (ret any, err exceptions.Exception) {
			writeMsg, ok := <-writeChannel
			if !ok {
				return nil, exceptions.NewRuntimeException(nil, "cannot receive msg from write channel", nil)
			}
			if writeErr := wsutil.WriteServerBinary(writeMsg.Conn, writeMsg.Data); writeErr != nil {
				err = exceptions.Package(writeErr)
				if writeMsg.ErrHandler != nil {
					writeMsg.ErrHandler(err)
					err = nil
				}
				return
			}
			return
		}, func(panic any) (ret any, err exceptions.Exception) {
			return nil, exceptions.PackagePanic(err, "an panic caused on handle websocket write")
		})
		exceptions.Print(err)
	}
}

func (a *AttachmentKey[T]) Name() string {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentKey is null", nil))
	}
	return a.name
}

func (a *AttachmentKey[T]) Id() int32 {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentKey is null", nil))
	}
	return a.id
}

func (a *AttachmentKey[T]) Get(c *AttachmentConn) *Attachment[T] {
	if c == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", nil))
	}
	return &Attachment[T]{
		key:        a,
		attachment: c.attachment,
	}
}

func (l *EventListener) Remove() exceptions.Exception {
	if l == nil {
		return nil
	}
	return l.node.Remove()
}

func (c *AttachmentConn) WriteChannel() chan<- ConnWriteMsg {
	return c.writeChannel
}

func (c *AttachmentConn) WriteData(data []byte) {
	c.writeChannel <- ConnWriteMsg{Conn: c, Data: data}
}

func NewAttachmentKey[T any](name string) AttachmentKey[T] {
	return AttachmentKey[T]{
		name: name,
		id:   atomic.AddInt32(&attachmentKeyId, 1),
	}
}

func NewSimpleAttachmentConn(conn net.Conn, writeChannel chan ConnWriteMsg) *AttachmentConn {
	return &AttachmentConn{
		conn:         conn,
		attachment:   new(sync.Map),
		writeChannel: writeChannel,
	}
}

//  syntax error: method must have no type parameters
//func (a *AttachmentConn) Get[T any](key *AttachmentKey[T]) *Attachment[T] {
//	if a == nil {
//		panic(exceptions.NewNPE("AttachmentConn is null", nil))
//	}
//	return &Attachment[T]{
//		key:        key,
//		attachment: a.attachment,
//	}
//}

func (c *AttachmentConn) notify(event ConnEvent) {
	if c == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", nil))
	}
	err := collections.Loop[*EventListener](&c.eventListenerList, func(element *EventListener) exceptions.Exception {
		_, _ = exceptions.Try(func() (ret any, err exceptions.Exception) {
			element.listener(event)
			return
		}, func(panic any) (ret any, err exceptions.Exception) {
			exceptions.NewRuntimeException(
				panic,
				"an exception caused on call ConnEvent listener:",
				exceptions.DefaultExceptionConfig().SetCause(panic),
			).PrintStackTrace()
			return
		})
		return nil
	})
	exceptions.Print(err)
}

func (c *AttachmentConn) AddEventListener(f func(event ConnEvent)) (listener *EventListener) {
	if c == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", nil))
	}
	if f != nil {
		listener = &EventListener{listener: f}
		listener.node = exceptions.Exec1r1(c.eventListenerList.OfferAndGetNode, listener)
	}
	return
}

func (c *AttachmentConn) Read(b []byte) (n int, err error) {
	if c == nil {
		return 0, exceptions.NewNPE("AttachmentConn is null", nil)
	}
	read, err := c.conn.Read(b)
	return read, exceptions.Package(err)
}

func (c *AttachmentConn) Write(b []byte) (n int, err error) {
	if c == nil {
		return 0, exceptions.NewNPE("AttachmentConn is null", nil)
	}
	write, err := c.conn.Write(b)
	return write, exceptions.Package(err)
}

func (c *AttachmentConn) Close() error {
	if c == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", nil))
	}
	go c.notify(NewConnClosed(c))
	return exceptions.Package(c.conn.Close())
}

func (c *AttachmentConn) LocalAddr() net.Addr {
	if c == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", nil))
	}
	return c.conn.LocalAddr()
}

func (c *AttachmentConn) RemoteAddr() net.Addr {
	if c == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", nil))
	}
	return c.conn.RemoteAddr()
}

func (c *AttachmentConn) SetDeadline(t time.Time) error {
	if c == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", nil))
	}
	return exceptions.Package(c.conn.SetDeadline(t))
}

func (c *AttachmentConn) SetReadDeadline(t time.Time) error {
	if c == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", nil))
	}
	return exceptions.Package(c.conn.SetReadDeadline(t))
}

func (c *AttachmentConn) SetWriteDeadline(t time.Time) error {
	if c == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", nil))
	}
	return exceptions.Package(c.conn.SetWriteDeadline(t))
}

func (a *Attachment[T]) Get() T {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", nil))
	}
	load, _ := a.attachment.Load(a.key.id)
	value, ok := load.(T)
	if !ok {
		panic(exception.NewTypeCastExceptionByType[T](load, nil))
	}
	return value
}

func (a *Attachment[T]) Set(value T) {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", nil))
	}
	a.attachment.Store(a.key.id, value)
}
