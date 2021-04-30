package context

import "tursom-im/im_conn"

type UserConnContext struct {
	connMap map[string]*im_conn.ConnGroup
}

func NewUserConnContext() *UserConnContext {
	return &UserConnContext{}
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

func (c *UserConnContext) AddUserConn(uid string, conn *im_conn.AttachmentConn) {
	c.TouchUserConn(uid).Add(conn)
}
