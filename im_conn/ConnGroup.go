package im_conn

import (
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/protobuf/proto"
	"github.com/tursom-im/tursom_im_protobuf"
	"github.com/tursom/GoCollections/concurrent"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
)

type void struct{}

var member void
var connGroupAttrKey = NewAttachmentKey[*EventListener]()

type ConnGroup struct {
	lang.BaseObject
	lock     concurrent.RWLock
	connList map[*AttachmentConn]void
	subGroup *ConnGroup
}

func NewConnGroup() *ConnGroup {
	return &ConnGroup{
		lock:     concurrent.NewReentrantRWLock(),
		connList: make(map[*AttachmentConn]void),
	}
}

func SnapshotConnGroup(g *ConnGroup) *ConnGroup {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	return &ConnGroup{
		lock:     g.lock,
		connList: g.connList,
	}
}

func (g *ConnGroup) Size() int32 {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	return int32(len(g.connList))
}

func (g *ConnGroup) Add(conn *AttachmentConn) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	if conn == nil {
		return
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	g.connList[conn] = member
	listener := conn.AddEventListener(g.connClosedListener)
	connGroupAttrKey.Get(conn).Set(listener)
}

func (g *ConnGroup) connClosedListener(i ConnEvent) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	if i.EventId().IsConnClosed() {
		g.Remove(i.Conn())
	}
}

func (g *ConnGroup) Remove(conn *AttachmentConn) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	_ = connGroupAttrKey.Get(conn).Get().Remove()
	g.lock.Lock()
	defer g.lock.Unlock()

	delete(g.connList, conn)
}

func (g *ConnGroup) WriteBinaryFrame(bytes []byte, filter func(*AttachmentConn) bool) int32 {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	var sent int32 = 0
	g.Loop(func(conn *AttachmentConn) {
		if filter == nil || filter(conn) {
			_, err := exceptions.Try(func() (ret any, err exceptions.Exception) {
				conn.WriteData(bytes)
				sent++
				return nil, nil
			}, func(panic any) (ret any, err exceptions.Exception) {
				return nil, exceptions.Package(err)
			})
			exceptions.Print(err)
		}
	})
	return sent
}

func (g *ConnGroup) WriteTextFrame(text string, filter func(*AttachmentConn) bool) int32 {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	var sent int32 = 0
	bytes := []byte(text)
	g.Loop(func(conn *AttachmentConn) {
		if filter == nil || filter(conn) {
			err := wsutil.WriteServerText(conn, bytes)
			if err != nil {
				exceptions.Print(err)
				g.Remove(conn)
				err = conn.Close()
				exceptions.Print(conn.Close())
			} else {
				sent++
			}
		}
	})
	return sent
}

func (g *ConnGroup) WriteChatMsg(msg *tursom_im_protobuf.ImMsg, filter func(*AttachmentConn) bool) exceptions.Exception {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
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
		panic(exceptions.NewNPE("ConnGroup is null", nil))
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
		panic(exceptions.NewNPE("ConnGroup is null", nil))
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
		panic(exceptions.NewNPE("ConnGroup is null", nil))
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
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	g.lock.RLock()
	defer g.lock.RUnlock()
	for conn := range g.connList {
		if conn != nil {
			handler(conn)
		}
	}
}
