package im_conn

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/tursom/GoCollections/collections"
	"github.com/tursom/GoCollections/exceptions"
	"net"
	"reflect"
	"sync"
	"time"
	"tursom-im/exception"
	"tursom-im/utils"
)

type AttachmentKey struct {
	name string
	t    reflect.Type
}

func (a *AttachmentKey) T() reflect.Type {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentKey is null", true))
	}
	return a.t
}

func (a *AttachmentKey) Name() string {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentKey is null", true))
	}
	return a.name
}

type Attachment struct {
	key        *AttachmentKey
	attachment *sync.Map
}

type readData struct {
	data []byte
	op   ws.OpCode
	err  error
}

type AttachmentConn struct {
	conn              net.Conn
	attachment        *sync.Map
	eventListenerList collections.MutableList
	writeChannel      chan []byte
	readChannel       chan readData
}

func (c *AttachmentConn) WriteChannel() chan<- []byte {
	return c.writeChannel
}

func (c *AttachmentConn) TryRead() ([]byte, ws.OpCode, error) {
	select {
	case data := <-c.readChannel:
		return data.data, data.op, data.err
	default:
	}
	return nil, 0, nil
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
		writeChannel:      make(chan []byte, 128),
		readChannel:       make(chan readData, 128),
	}
}

func NewSimpleAttachmentConn(conn net.Conn) *AttachmentConn {
	var attachment sync.Map
	return &AttachmentConn{
		conn:              conn,
		attachment:        &attachment,
		eventListenerList: collections.NewArrayList(),
		writeChannel:      make(chan []byte, 128),
		readChannel:       make(chan readData, 128),
	}
}
func (a *AttachmentConn) LoopRead() {
	var err error
	for !utils.IsClosedError(err) {
		_, err = exceptions.Try(func() (ret interface{}, err exceptions.Exception) {
			msg, op, e := wsutil.ReadClientData(a)
			if e != nil {
				return nil, exceptions.Package(e)
			}
			a.readChannel <- readData{
				data: msg,
				op:   op,
			}
			return nil, nil
		}, func(panic interface{}) (ret interface{}, err exceptions.Exception) {
			return nil, exceptions.Package(err)
		})
		a.readChannel <- readData{
			err: exceptions.Package(err),
		}
	}
}

func (a *AttachmentConn) Get(key *AttachmentKey) *Attachment {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", true))
	}
	return &Attachment{
		key:        key,
		attachment: a.attachment,
	}
}

func (a *AttachmentConn) notify(event ConnEvent) {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", true))
	}
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
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", true))
	}
	load, _ := a.attachment.Load(a.key.name)
	return load
}

func (a *Attachment) Set(value interface{}) exceptions.Exception {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", true))
	}
	valueType := reflect.TypeOf(value)
	if valueType.AssignableTo(a.key.t) {
		a.attachment.Store(a.key.name, value)
		return nil
	} else {
		return exception.NewInvalidTypeException(fmt.Sprintf("value of type %s cannot cast to %s", valueType, a.key.t))
	}
}

func (a *AttachmentConn) AddEventListener(f func(event ConnEvent)) {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", true))
	}
	if f != nil {
		a.eventListenerList.Add(f)
	}
}

func (a *AttachmentConn) RemoveEventListener(f func(ConnEvent)) {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", true))
	}
	if f != nil {
		_ = a.eventListenerList.Remove(f)
	}
}

func (a *AttachmentConn) Read(b []byte) (n int, err error) {
	if a == nil {
		return 0, exceptions.NewNPE("AttachmentConn is null", true)
	}
	read, err := a.conn.Read(b)
	return read, exceptions.Package(err)
}

func (a *AttachmentConn) Write(b []byte) (n int, err error) {
	if a == nil {
		return 0, exceptions.NewNPE("AttachmentConn is null", true)
	}
	write, err := a.conn.Write(b)
	return write, exceptions.Package(err)
}

func (a *AttachmentConn) HandleWrite() error {
	for true {
		select {
		case bytes := <-a.writeChannel:
			err := wsutil.WriteServerBinary(a, bytes)
			if err != nil {
				return err
			}
		default:
			return nil
		}
	}
	return nil
}

func (a *AttachmentConn) Close() error {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", true))
	}
	a.notify(NewConnClosed(a))
	return exceptions.Package(a.conn.Close())
}

func (a *AttachmentConn) LocalAddr() net.Addr {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", true))
	}
	return a.conn.LocalAddr()
}

func (a *AttachmentConn) RemoteAddr() net.Addr {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", true))
	}
	return a.conn.RemoteAddr()
}

func (a *AttachmentConn) SetDeadline(t time.Time) error {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", true))
	}
	return exceptions.Package(a.conn.SetDeadline(t))
}

func (a *AttachmentConn) SetReadDeadline(t time.Time) error {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", true))
	}
	return exceptions.Package(a.conn.SetReadDeadline(t))
}

func (a *AttachmentConn) SetWriteDeadline(t time.Time) error {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentConn is null", true))
	}
	return exceptions.Package(a.conn.SetWriteDeadline(t))
}
