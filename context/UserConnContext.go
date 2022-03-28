package context

import (
	"github.com/tursom-im/im_conn"
	"github.com/tursom/GoCollections/concurrent"
	"github.com/tursom/GoCollections/lang"
)

type UserConnContext struct {
	lang.BaseObject
	connMap     map[string]*im_conn.ConnGroup
	attrContext *AttrContext
	lock        concurrent.RWLock
}

func NewUserConnContext() *UserConnContext {
	return &UserConnContext{connMap: make(map[string]*im_conn.ConnGroup), attrContext: nil}
}

func (c UserConnContext) Init(ctx *GlobalContext) {
	c.attrContext = ctx.attrContext
}

func (c *UserConnContext) TouchUserConn(uid string) *im_conn.ConnGroup {
	c.lock.Lock()
	defer c.lock.Unlock()

	group := c.connMap[uid]
	if group == nil {
		group = im_conn.NewConnGroup()
		c.connMap[uid] = group
	}
	return group
}

func (c *UserConnContext) GetUserConn(uid string) *im_conn.ConnGroup {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.connMap[uid]
}

func (c *UserConnContext) RemoveUserConn(uid string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.connMap, uid)
}

func (c *UserConnContext) GetCurrentConn(conn *im_conn.AttachmentConn) *im_conn.ConnGroup {
	currentUserId := c.attrContext.userIdAttrKey.Get(conn)
	return c.GetUserConn(currentUserId.Get().AsString())
}

func (c *UserConnContext) AddUserConn(uid string, conn *im_conn.AttachmentConn) {
	c.TouchUserConn(uid).Add(conn)
}
