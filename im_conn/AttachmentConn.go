package im_conn

import (
	"fmt"
	"github.com/tursom/GoCollections/collections"
	"github.com/tursom/GoCollections/exceptions"
	"net"
	"reflect"
	"sync"
	"time"
	"tursom-im/exception"
)

type AttachmentKey struct {
	name string
	t    reflect.Type
}

func (a *AttachmentKey) T() reflect.Type {
	return a.t
}

func (a *AttachmentKey) Name() string {
	return a.name
}

type Attachment struct {
	key        *AttachmentKey
	attachment *sync.Map
}

type AttachmentConn struct {
	conn              net.Conn
	attachment        *sync.Map
	eventListenerList collections.MutableList
}

type InvalidTypeError struct {
}

func (i *InvalidTypeError) Error() string {
	return "invalid type"
}

func NewAttachmentKey(name string, t reflect.Type) AttachmentKey {
	return AttachmentKey{
		name: name,
		t:    t,
	}
}

func NewAttachmentConn(conn net.Conn, attachment *sync.Map) *AttachmentConn {
	if attachment != nil {
		var newMap sync.Map
		attachment = &newMap
	}
	return &AttachmentConn{
		conn:              conn,
		attachment:        attachment,
		eventListenerList: collections.NewArrayList(),
	}
}

func NewSimpleAttachmentConn(conn net.Conn) *AttachmentConn {
	var attachment sync.Map
	return &AttachmentConn{
		conn:              conn,
		attachment:        &attachment,
		eventListenerList: collections.NewArrayList(),
	}
}

func (a *AttachmentConn) Get(key *AttachmentKey) *Attachment {
	return &Attachment{
		key:        key,
		attachment: a.attachment,
	}
}

func (a *AttachmentConn) notify(event ConnEvent) {
	err := collections.Loop(a.eventListenerList, func(element interface{}) exceptions.Exception {
		switch element.(type) {
		case func(ConnEvent):
			_, _ = exceptions.Try(func() (ret interface{}, err exceptions.Exception) {
				element.(func(ConnEvent))(event)
				return
			}, func(panic interface{}) (ret interface{}, err exceptions.Exception) {
				exceptions.NewRuntimeException(
					panic,
					"an exception caused on call ConnEvent listener:",
					true, panic,
				).PrintStackTrace()
				return
			})
		}
		return nil
	})
	exceptions.Print(err)
}

func (a *Attachment) Get() interface{} {
	load, _ := a.attachment.Load(a.key.name)
	return load
}

func (a *Attachment) Set(value interface{}) exceptions.Exception {
	valueType := reflect.TypeOf(value)
	if valueType.AssignableTo(a.key.t) {
		a.attachment.Store(a.key.name, value)
		return nil
	} else {
		return exception.NewInvalidTypeException(fmt.Sprintf("value of type %s cannot cast to %s", valueType, a.key.t))
	}
}

func (a *AttachmentConn) AddEventListener(f func(event ConnEvent)) {
	if f != nil {
		a.eventListenerList.Add(f)
	}
}

func (a *AttachmentConn) RemoveEventListener(f func(ConnEvent)) {
	if f != nil {
		_ = a.eventListenerList.Remove(f)
	}
}

func (a *AttachmentConn) Read(b []byte) (n int, err error) {
	read, err := a.conn.Read(b)
	return read, exceptions.Package(err)
}

func (a *AttachmentConn) Write(b []byte) (n int, err error) {
	write, err := a.conn.Write(b)
	return write, exceptions.Package(err)
}

func (a *AttachmentConn) Close() error {
	a.notify(NewConnClosed(a))
	return exceptions.Package(a.conn.Close())
}

func (a *AttachmentConn) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *AttachmentConn) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *AttachmentConn) SetDeadline(t time.Time) error {
	return exceptions.Package(a.conn.SetDeadline(t))
}

func (a *AttachmentConn) SetReadDeadline(t time.Time) error {
	return exceptions.Package(a.conn.SetReadDeadline(t))
}

func (a *AttachmentConn) SetWriteDeadline(t time.Time) error {
	return exceptions.Package(a.conn.SetWriteDeadline(t))
}
