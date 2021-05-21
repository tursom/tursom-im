package im_conn

type EventId int32

const (
	ConnClosedId = 1
)

type ConnEvent interface {
	EventId() EventId
	Conn() *AttachmentConn
}

type AbstractConnEvent struct {
	eventId EventId
	conn    *AttachmentConn
}

func (a AbstractConnEvent) EventId() EventId {
	return a.eventId
}

func (a AbstractConnEvent) Conn() *AttachmentConn {
	return a.conn
}

func NewAbstractConnEvent(eventId EventId, conn *AttachmentConn) AbstractConnEvent {
	return AbstractConnEvent{eventId: eventId, conn: conn}
}

type ConnClosed struct {
	AbstractConnEvent
}

func NewConnClosed(conn *AttachmentConn) ConnClosed {
	return ConnClosed{NewAbstractConnEvent(ConnClosedId, conn)}
}
