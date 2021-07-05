package context

import (
	"math/rand"
	"sync"
	"tursom-im/im_conn"
)

// ConnNodeContext
// 负责节点注册的服务
type ConnNodeContext struct {
	connMap     map[int32]*im_conn.AttachmentConn
	mutex       *sync.Mutex
	attrContext *AttrContext
	nodeMax     int32
}

func NewConnNodeContext(nodeMax int32) *ConnNodeContext {
	return &ConnNodeContext{
		connMap:     make(map[int32]*im_conn.AttachmentConn),
		mutex:       new(sync.Mutex),
		attrContext: nil,
		nodeMax:     nodeMax,
	}
}

func (c *ConnNodeContext) Init(ctx *GlobalContext) {
	c.attrContext = ctx.attrContext
}

func (c *ConnNodeContext) Allocate(conn *im_conn.AttachmentConn) int32 {
	if conn == nil {
		return -1
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	randNode := rand.Int31()
	if c.check(randNode) {
		c.register(randNode, conn)
		return randNode
	}

	node := randNode - 1
	for node >= 0 {
		if c.check(node) {
			c.register(node, conn)
			return node
		}
		node--
	}

	node = randNode + 1
	for node < c.nodeMax {
		if c.check(node) {
			c.register(node, conn)
			return node
		}
		node++
	}

	return -1
}

func (c *ConnNodeContext) check(node int32) bool {
	return c.connMap[node] == nil
}

func (c *ConnNodeContext) register(node int32, conn *im_conn.AttachmentConn) {
	c.connMap[node] = conn
	conn.AddEventListener(func(event im_conn.ConnEvent) {
		switch event.EventId() {
		case im_conn.ConnClosedId:
			c.mutex.Lock()
			defer c.mutex.Unlock()

			conn := c.connMap[node]
			if conn == event.Conn() {
				delete(c.connMap, node)
			}
		}
	})
}
