package im_conn

import (
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/protobuf/proto"
	"github.com/tursom/GoCollections/exceptions"
	"tursom-im/tursom_im_protobuf"
)

type void struct{}

var member void

type ConnGroup struct {
	connList map[*AttachmentConn]void
}

func NewConnGroup() *ConnGroup {
	return &ConnGroup{
		connList: make(map[*AttachmentConn]void),
	}
}

func (g *ConnGroup) Add(conn *AttachmentConn) {
	if conn == nil {
		return
	}
	g.connList[conn] = member
	conn.AddEventListener(g.connClosedListener)
}

func (g *ConnGroup) connClosedListener(i ConnEvent) {
	switch i.EventId() {
	case ConnClosedId:
		g.Remove(i.Conn())
	}
}

func (g *ConnGroup) Remove(conn *AttachmentConn) {
	delete(g.connList, conn)
}

func (g *ConnGroup) WriteBinaryFrame(bytes []byte, filter func(*AttachmentConn) bool) {
	for conn := range g.connList {
		if filter == nil || filter(conn) {
			err := wsutil.WriteServerBinary(conn, bytes)
			if err != nil {
				exceptions.Print(err)
				exceptions.Print(conn.Close())
				g.Remove(conn)
			}
		}
	}
}

func (g *ConnGroup) WriteTextFrame(text string, filter func(*AttachmentConn) bool) {
	bytes := []byte(text)
	for conn := range g.connList {
		if filter == nil || filter(conn) {
			err := wsutil.WriteServerText(conn, bytes)
			if err != nil {
				exceptions.Print(err)
				err = conn.Close()
				exceptions.Print(conn.Close())
				g.Remove(conn)
			}
		}
	}
}

func (g *ConnGroup) WriteChatMsg(msg *tursom_im_protobuf.ImMsg, filter func(*AttachmentConn) bool) exceptions.Exception {
	bytes, err := proto.Marshal(msg)
	if err != nil {
		return exceptions.Package(err)
	}
	g.WriteBinaryFrame(bytes, filter)
	return nil
}
