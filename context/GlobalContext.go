package context

import (
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"

	"github.com/tursom-im/config"
)

type GlobalContext struct {
	lang.BaseObject
	cfg              *config.Config
	tokenContext     *TokenContext
	attrContext      *AttrContext
	userConnContext  *UserConnContext
	msgIdContext     *MsgIdContext
	sqlContext       SqlContext
	connNodeContext  *ConnNodeContext
	broadcastContext BroadcastContext
}

func NewGlobalContext(config *config.Config) *GlobalContext {
	sqlContext := NewSqliteSqlContext()
	if sqlContext == nil {
		return nil
	}

	tokenContext := NewTokenContext()
	if tokenContext == nil {
		return nil
	}

	attrContext := NewAttrContext()
	if attrContext == nil {
		return nil
	}

	userConnContext := NewUserConnContext()
	if userConnContext == nil {
		return nil
	}

	msgIdContext := NewMsgIdContext()
	if msgIdContext == nil {
		return nil
	}

	connNodeContext := NewConnNodeContext(config.Node.NodeMax)
	if connNodeContext == nil {
		return nil
	}

	broadcastContext := NewBroadcastContext()
	if broadcastContext == nil {
		return nil
	}

	g := &GlobalContext{
		cfg:              config,
		tokenContext:     tokenContext,
		attrContext:      attrContext,
		userConnContext:  userConnContext,
		msgIdContext:     msgIdContext,
		sqlContext:       sqlContext,
		connNodeContext:  connNodeContext,
		broadcastContext: broadcastContext,
	}
	g.tokenContext.Init(g)
	g.userConnContext.Init(g)
	g.sqlContext.Init(g)
	connNodeContext.Init(g)
	return g
}

func (g *GlobalContext) Cfg() *config.Config {
	return g.Config()
}

func (g *GlobalContext) Config() *config.Config {
	exceptions.CheckNil(g)
	return g.cfg
}

func (g *GlobalContext) Attr() *AttrContext {
	return g.AttrContext()
}

func (g *GlobalContext) AttrContext() *AttrContext {
	exceptions.CheckNil(g)
	return g.attrContext
}

func (g *GlobalContext) Token() *TokenContext {
	return g.TokenContext()
}

func (g *GlobalContext) TokenContext() *TokenContext {
	exceptions.CheckNil(g)
	return g.tokenContext
}

func (g *GlobalContext) UserConn() *UserConnContext {
	return g.UserConnContext()
}

func (g *GlobalContext) UserConnContext() *UserConnContext {
	exceptions.CheckNil(g)
	return g.userConnContext
}

func (g *GlobalContext) MsgId() *MsgIdContext {
	return g.MsgIdContext()
}

func (g *GlobalContext) MsgIdContext() *MsgIdContext {
	exceptions.CheckNil(g)
	return g.msgIdContext
}

func (g *GlobalContext) Sql() SqlContext {
	return g.SqlContext()
}

func (g *GlobalContext) SqlContext() SqlContext {
	exceptions.CheckNil(g)
	return g.sqlContext
}

func (g *GlobalContext) ConnNode() *ConnNodeContext {
	return g.ConnNodeContext()
}

func (g *GlobalContext) ConnNodeContext() *ConnNodeContext {
	exceptions.CheckNil(g)
	return g.connNodeContext
}

func (g *GlobalContext) Broadcast() BroadcastContext {
	return g.BroadcastContext()
}

func (g *GlobalContext) BroadcastContext() BroadcastContext {
	exceptions.CheckNil(g)
	return g.broadcastContext
}
