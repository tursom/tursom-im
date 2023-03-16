package conn

import (
	"github.com/tursom/GoCollections/concurrent"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"

	"github.com/tursom-im/proto/pkg"
)

type ConnGroup struct {
	lang.BaseObject
	lock     concurrent.RWLock
	connMap  map[Conn]EventListener
	subGroup *ConnGroup
}

func NewConnGroup() *ConnGroup {
	return &ConnGroup{
		lock:    concurrent.NewReentrantRWLock(),
		connMap: make(map[Conn]EventListener),
	}
}

func SnapshotConnGroup(g *ConnGroup) *ConnGroup {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	return &ConnGroup{
		lock:    g.lock,
		connMap: g.connMap,
	}
}

func (g *ConnGroup) Size() int32 {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	return int32(len(g.connMap))
}

func (g *ConnGroup) Add(conn Conn) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	if conn == nil {
		return
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	if _, ok := g.connMap[conn]; ok {
		return
	}
	g.connMap[conn] = conn.AddEventListener(g.connClosedListener)
}

func (g *ConnGroup) connClosedListener(i Event) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	if i.EventId().IsConnClosed() {
		g.Remove(i.Conn())
	}
}

func (g *ConnGroup) Remove(conn Conn) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	_ = g.connMap[conn].Remove()
	delete(g.connMap, conn)
}

func (g *ConnGroup) WriteChatMsg(msg *pkg.ImMsg, filter func(Conn) bool) int {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}

	var sent = 0
	g.Loop(func(conn Conn) {
		if filter == nil || filter(conn) {
			defer func() {
				exceptions.Print(exceptions.PackageAny(recover()))
			}()
			conn.SendMsg(msg)
			sent++
		}
	})
	return sent
}

func (g *ConnGroup) Append(target *ConnGroup) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	if target == nil {
		return
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	target.Loop(func(conn Conn) {
		if _, ok := g.connMap[conn]; ok {
			return
		}
		g.connMap[conn] = conn.AddEventListener(g.connClosedListener)
	})
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

func (g *ConnGroup) Loop(handler func(Conn)) {
	if g == nil {
		panic(exceptions.NewNPE("ConnGroup is null", nil))
	}
	g.lock.RLock()
	defer g.lock.RUnlock()
	for conn := range g.connMap {
		if conn != nil {
			handler(conn)
		}
	}
}
