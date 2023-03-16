package context

import (
	"math/rand"
	"sync"

	"github.com/tursom/GoCollections/concurrent"
	"github.com/tursom/GoCollections/lang"

	"github.com/tursom-im/conn"
)

// ConnNodeContext
// 负责节点注册的服务
type ConnNodeContext struct {
	lang.BaseObject
	connMap     map[int32]conn.Conn
	mutex       concurrent.Lock
	attrContext *AttrContext
	nodeMax     int32
}

func NewConnNodeContext(nodeMax int32) *ConnNodeContext {
	return &ConnNodeContext{
		connMap:     make(map[int32]conn.Conn),
		mutex:       new(sync.Mutex),
		attrContext: nil,
		nodeMax:     nodeMax,
	}
}

func (ctx *ConnNodeContext) Init(globalCtx *GlobalContext) {
	ctx.attrContext = globalCtx.attrContext
}

func (ctx *ConnNodeContext) Allocate(conn conn.Conn) int32 {
	if conn == nil {
		return -1
	}

	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	randNode := rand.Int31() % ctx.nodeMax
	if ctx.check(randNode) {
		ctx.register(randNode, conn)
		return randNode
	}

	node := randNode - 1
	for node >= 0 {
		if ctx.check(node) {
			ctx.register(node, conn)
			return node
		}
		node--
	}

	node = randNode + 1
	for node < ctx.nodeMax {
		if ctx.check(node) {
			ctx.register(node, conn)
			return node
		}
		node++
	}

	return -1
}

func (ctx *ConnNodeContext) check(node int32) bool {
	return ctx.connMap[node] == nil
}

func (ctx *ConnNodeContext) register(node int32, c conn.Conn) {
	ctx.connMap[node] = c
	c.AddEventListener(func(event conn.Event) {
		if !event.EventId().IsConnClosed() {
			return
		}
		ctx.mutex.Lock()
		defer ctx.mutex.Unlock()

		if ctx.connMap[node] == event.Conn() {
			delete(ctx.connMap, node)
		}
	})
}
