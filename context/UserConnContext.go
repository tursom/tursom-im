package context

import "tursom-im/im_conn"

type UserConnContext struct {
	connMap     map[string]*im_conn.ConnGroup
	attrContext *AttrContext
}

func NewUserConnContext() *UserConnContext {
	return &UserConnContext{make(map[string]*im_conn.ConnGroup), nil}
}

func (c UserConnContext) Init(ctx *GlobalContext) {
	c.attrContext = ctx.attrContext
}

func (c *UserConnContext) TouchUserConn(uid string) *im_conn.ConnGroup {
	group := c.connMap[uid]
	if group == nil {
		group = im_conn.NewConnGroup()
		c.connMap[uid] = group
	}
	return group
}

func (c *UserConnContext) GetUserConn(uid string) *im_conn.ConnGroup {
	return c.connMap[uid]
}

func (c *UserConnContext) RemoveUserConn(uid string) {
	delete(c.connMap, uid)
}

func (c *UserConnContext) GetCurrentConn(conn *im_conn.AttachmentConn) *im_conn.ConnGroup {
	currentUserId := conn.Get(c.attrContext.userIdAttrKey)
	return c.GetUserConn(currentUserId.Get().(string))
}

func (c *UserConnContext) AddUserConn(uid string, conn *im_conn.AttachmentConn) {
	c.TouchUserConn(uid).Add(conn)
}
