package im_conn

import "github.com/tursom/GoCollections/lang"

type EventId int32

const (
	ConnClosedId EventId = 1
)

type (
	ConnEvent interface {
		lang.Object
		EventId() EventId
		Conn() *AttachmentConn
	}

	AbstractConnEvent struct {
		lang.BaseObject
		eventId EventId
		conn    *AttachmentConn
	}

	ConnClosed struct {
		AbstractConnEvent
	}
)

func (a *AbstractConnEvent) EventId() EventId {
	return a.eventId
}

func (a *AbstractConnEvent) Conn() *AttachmentConn {
	return a.conn
}

func NewAbstractConnEvent(eventId EventId, conn *AttachmentConn) *AbstractConnEvent {
	return &AbstractConnEvent{eventId: eventId, conn: conn}
}

func NewConnClosed(conn *AttachmentConn) *ConnClosed {
	return &ConnClosed{AbstractConnEvent: *NewAbstractConnEvent(ConnClosedId, conn)}
}

func (i EventId) IsConnClosed() bool {
	return i == ConnClosedId
}
