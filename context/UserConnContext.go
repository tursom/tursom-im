package context

import (
	"sync"

	"github.com/tursom/GoCollections/concurrent"
	"github.com/tursom/GoCollections/lang"

	"github.com/tursom-im/conn"
)

type UserConnContext struct {
	lang.BaseObject
	connMap     map[string]*conn.ConnGroup
	attrContext *AttrContext
	lock        concurrent.RWLock
}

func NewUserConnContext() *UserConnContext {
	return &UserConnContext{connMap: make(map[string]*conn.ConnGroup), lock: &sync.RWMutex{}}
}

func (c *UserConnContext) Init(ctx *GlobalContext) {
	c.attrContext = ctx.attrContext
}

func (c *UserConnContext) TouchUserConn(uid string) *conn.ConnGroup {
	c.lock.Lock()
	defer c.lock.Unlock()

	group := c.connMap[uid]
	if group == nil {
		group = conn.NewConnGroup()
		c.connMap[uid] = group
	}
	return group
}

func (c *UserConnContext) GetUserConn(uid string) *conn.ConnGroup {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.connMap[uid]
}

func (c *UserConnContext) RemoveUserConn(uid string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.connMap, uid)
}

func (c *UserConnContext) GetCurrentConn(conn conn.Conn) *conn.ConnGroup {
	currentUserId := c.attrContext.userIdAttrKey.Get(conn)
	return c.GetUserConn(currentUserId.Get().AsString())
}

func (c *UserConnContext) AddUserConn(uid string, conn conn.Conn) {
	c.TouchUserConn(uid).Add(conn)
}
