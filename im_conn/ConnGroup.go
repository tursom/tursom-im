package im_conn

import (
	"github.com/gobwas/ws/wsutil"
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

func (g *ConnGroup) WriteBinaryFrame(bytes []byte) {
	for conn := range g.connList {
		wsutil.WriteServerBinary(conn, bytes)
	}
}

func (g *ConnGroup) WriteTextFrame(text string) {
	bytes := []byte(text)
	for conn := range g.connList {
		wsutil.WriteServerText(conn, bytes)
	}
}
