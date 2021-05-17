package im_conn

import (
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/protobuf/proto"
	tursom_im_protobuf "tursom-im/proto"
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
	g.connList[conn] = member
}

func (g *ConnGroup) Remove(conn *AttachmentConn) {
	delete(g.connList, conn)
}

func (g *ConnGroup) WriteBinaryFrame(bytes []byte, filter func(*AttachmentConn) bool) {
	for conn := range g.connList {
		if filter != nil && filter(conn) {
			wsutil.WriteServerBinary(conn, bytes)
		}
	}
}

func (g *ConnGroup) WriteTextFrame(text string, filter func(*AttachmentConn) bool) {
	bytes := []byte(text)
	for conn := range g.connList {
		if filter != nil && filter(conn) {
			wsutil.WriteServerText(conn, bytes)
		}
	}
}

func (g *ConnGroup) WriteChatMsg(msg *tursom_im_protobuf.ImMsg, filter func(*AttachmentConn) bool) error {
	bytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	g.WriteBinaryFrame(bytes, filter)
	return nil
}
