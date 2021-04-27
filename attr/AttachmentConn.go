package attr

import (
	"net"
	"reflect"
	"sync"
	"time"
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
	conn       *net.Conn
	attachment *sync.Map
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

func NewAttachmentConn(conn *net.Conn, attachment *sync.Map) *AttachmentConn {
	if attachment != nil {
		var newMap sync.Map
		attachment = &newMap
	}
	return &AttachmentConn{
		conn:       conn,
		attachment: attachment,
	}
}

func NewSimpleAttachmentConn(conn *net.Conn) *AttachmentConn {
	var attachment sync.Map
	return &AttachmentConn{
		conn:       conn,
		attachment: &attachment,
	}
}

func (a *AttachmentConn) Get(key *AttachmentKey) Attachment {
	return Attachment{
		key:        key,
		attachment: a.attachment,
	}
}

func (a *Attachment) Get() interface{} {
	load, _ := a.attachment.Load(a.key.name)
	return load
}

func (a *Attachment) Set(value interface{}) error {
	valueType := reflect.TypeOf(value)
	if valueType.AssignableTo(a.key.t) {
		a.attachment.Store(a.key.name, value)
		return nil
	} else {
		return &InvalidTypeError{}
	}
}

func (a *AttachmentConn) Read(b []byte) (n int, err error) {
	return (*a.conn).Read(b)
}

func (a *AttachmentConn) Write(b []byte) (n int, err error) {
	return (*a.conn).Write(b)
}

func (a *AttachmentConn) Close() error {
	return (*a.conn).Close()
}

func (a *AttachmentConn) LocalAddr() net.Addr {
	return (*a.conn).LocalAddr()
}

func (a *AttachmentConn) RemoteAddr() net.Addr {
	return (*a.conn).RemoteAddr()
}

func (a *AttachmentConn) SetDeadline(t time.Time) error {
	return (*a.conn).SetDeadline(t)
}

func (a *AttachmentConn) SetReadDeadline(t time.Time) error {
	return (*a.conn).SetReadDeadline(t)
}

func (a *AttachmentConn) SetWriteDeadline(t time.Time) error {
	return (*a.conn).SetWriteDeadline(t)
}
