package im_conn

import (
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/protobuf/proto"
	"github.com/tursom/GoCollections/exceptions"
	"sync"
	"tursom-im/tursom_im_protobuf"
	"tursom-im/utils"
)

type void struct{}

var member void

type ConnGroup struct {
	lock     *sync.RWMutex
	connList map[*AttachmentConn]void
}

func NewConnGroup() *ConnGroup {
	return &ConnGroup{
		lock:     new(sync.RWMutex),
		connList: make(map[*AttachmentConn]void),
	}
}

func (g *ConnGroup) Size() int32 {
	return int32(len(g.connList))
}

func (g *ConnGroup) Add(conn *AttachmentConn) {
	if conn == nil {
		return
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	g.connList[conn] = member
	conn.AddEventListener(g.connClosedListener)
}

func (g *ConnGroup) connClosedListener(i ConnEvent) {
	if i.EventId().IsConnClosed() {
		g.Remove(i.Conn())
	}
}

func (g *ConnGroup) Remove(conn *AttachmentConn) {
	g.lock.Lock()
	defer g.lock.Unlock()
	delete(g.connList, conn)
}

func (g *ConnGroup) WriteBinaryFrame(bytes []byte, filter func(*AttachmentConn) bool) int32 {
	var sent int32 = 0
	g.lock.RLock()
	defer g.lock.RUnlock()
	for conn := range g.connList {
		if filter == nil || filter(conn) {
			err := wsutil.WriteServerBinary(conn, bytes)
			if err != nil {
				if !utils.IsClosedError(err) {
					exceptions.Print(err)
					exceptions.Print(conn.Close())
				}
				g.Remove(conn)
			} else {
				sent++
			}
		}
	}
	return sent
}

func (g *ConnGroup) WriteTextFrame(text string, filter func(*AttachmentConn) bool) int32 {
	var sent int32 = 0
	bytes := []byte(text)
	g.lock.RLock()
	defer g.lock.RUnlock()
	for conn := range g.connList {
		if filter == nil || filter(conn) {
			err := wsutil.WriteServerText(conn, bytes)
			if err != nil {
				if !utils.IsClosedError(err) {
					exceptions.Print(err)
				}
				err = conn.Close()
				exceptions.Print(conn.Close())
				g.Remove(conn)
			} else {
				sent++
			}
		}
	}
	return sent
}

func (g *ConnGroup) WriteChatMsg(msg *tursom_im_protobuf.ImMsg, filter func(*AttachmentConn) bool) exceptions.Exception {
	bytes, err := proto.Marshal(msg)
	if err != nil {
		return exceptions.Package(err)
	}
	g.WriteBinaryFrame(bytes, filter)
	return nil
}
