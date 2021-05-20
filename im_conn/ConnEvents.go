package im_conn

type EventId int32

const (
	ConnClosedId = 1
)

type ConnEvent interface {
	EventId() EventId
	Conn() *AttachmentConn
}

type ConnClosed struct {
	conn *AttachmentConn
}

func (c ConnClosed) EventId() EventId {
	return ConnClosedId
}

func (c ConnClosed) Conn() *AttachmentConn {
	return c.conn
}
