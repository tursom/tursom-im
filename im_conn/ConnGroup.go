package im_conn

import (
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/protobuf/proto"
	"github.com/tursom/GoCollections/exceptions"
	"sync"
	"tursom-im/tursom_im_protobuf"
)

type void struct{}

var member void

type ConnGroup struct {
	lock     *sync.RWMutex
	connList map[*AttachmentConn]void
	subGroup *ConnGroup
}

func NewConnGroup() *ConnGroup {
	return &ConnGroup{
		lock:     new(sync.RWMutex),
		connList: make(map[*AttachmentConn]void),
	}
}

func SnapshotConnGroup(g *ConnGroup) *ConnGroup {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", true))
	}
	return &ConnGroup{
		lock:     g.lock,
		connList: g.connList,
	}
}

func (g *ConnGroup) Size() int32 {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", true))
	}
	return int32(len(g.connList))
}

func (g *ConnGroup) Add(conn *AttachmentConn) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", true))
	}
	if conn == nil {
		return
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	g.connList[conn] = member
	conn.AddEventListener(g.connClosedListener)
}

func (g *ConnGroup) connClosedListener(i ConnEvent) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", true))
	}
	if i.EventId().IsConnClosed() {
		g.Remove(i.Conn())
	}
}

func (g *ConnGroup) Remove(conn *AttachmentConn) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", true))
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	delete(g.connList, conn)
}

func (g *ConnGroup) WriteBinaryFrame(bytes []byte, filter func(*AttachmentConn) bool) int32 {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", true))
	}
	var sent int32 = 0
	g.lock.RLock()
	defer g.lock.RUnlock()
	g.Loop(func(conn *AttachmentConn) {
		if filter == nil || filter(conn) {
			_, err := exceptions.Try(func() (ret interface{}, err exceptions.Exception) {
				conn.WriteChannel() <- bytes
				sent++
				return nil, nil
			}, func(panic interface{}) (ret interface{}, err exceptions.Exception) {
				return nil, exceptions.Package(err)
			})
			exceptions.Print(err)
		}
	})
	return sent
}

func (g *ConnGroup) WriteTextFrame(text string, filter func(*AttachmentConn) bool) int32 {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", true))
	}
	var sent int32 = 0
	bytes := []byte(text)
	g.lock.RLock()
	defer g.lock.RUnlock()
	g.Loop(func(conn *AttachmentConn) {
		if filter == nil || filter(conn) {
			err := wsutil.WriteServerText(conn, bytes)
			if err != nil {
				exceptions.Print(err)
				err = conn.Close()
				exceptions.Print(conn.Close())
				g.Remove(conn)
			} else {
				sent++
			}
		}
	})
	return sent
}

func (g *ConnGroup) WriteChatMsg(msg *tursom_im_protobuf.ImMsg, filter func(*AttachmentConn) bool) exceptions.Exception {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", true))
	}
	bytes, err := proto.Marshal(msg)
	if err != nil {
		return exceptions.Package(err)
	}
	g.WriteBinaryFrame(bytes, filter)
	return nil
}

func (g *ConnGroup) Append(target *ConnGroup) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", true))
	}
	if target == nil {
		return
	}
	target.Loop(func(conn *AttachmentConn) {
		g.Add(conn)
	})
}

func (g *ConnGroup) Link(target *ConnGroup) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", true))
	}
	defer g.lock.Unlock()
	g.lock.Lock()

	subGroup := g
	for subGroup.subGroup != nil {
		subGroup = subGroup.subGroup
	}
	subGroup.subGroup = target
}

func (g *ConnGroup) Aggregation(target *ConnGroup) *ConnGroup {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", true))
	}
	if g == target {
		return g
	}
	group := NewConnGroup()
	group.Append(g)
	group.Append(target)
	return group
}

func (g *ConnGroup) Loop(handler func(*AttachmentConn)) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", true))
	}
	g.lock.RLock()
	defer func() {
		g.lock.RUnlock()
		if g == nil {
			g.subGroup.Loop(handler)
		}
	}()
	for conn := range g.connList {
		if conn != nil {
			handler(conn)
		}
	}
}
