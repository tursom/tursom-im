package context

import (
	"github.com/tursom-im/config"
	"github.com/tursom/GoCollections/lang"
)

type GlobalContext struct {
	lang.BaseObject
	tokenContext     *TokenContext
	attrContext      *AttrContext
	userConnContext  *UserConnContext
	msgIdContext     *MsgIdContext
	cfg              *config.Config
	sqlContext       SqlContext
	connNodeContext  *ConnNodeContext
	broadcastContext *BroadcastContext
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
		tokenContext:     tokenContext,
		attrContext:      attrContext,
		userConnContext:  userConnContext,
		msgIdContext:     msgIdContext,
		cfg:              config,
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

func (g *GlobalContext) Config() *config.Config {
	return g.cfg
}

func (g *GlobalContext) AttrContext() *AttrContext {
	return g.attrContext
}

func (g *GlobalContext) TokenContext() *TokenContext {
	return g.tokenContext
}

func (g *GlobalContext) UserConnContext() *UserConnContext {
	return g.userConnContext
}

func (g *GlobalContext) MsgIdContext() *MsgIdContext {
	return g.msgIdContext
}

func (g *GlobalContext) Cfg() *config.Config {
	return g.cfg
}

func (g *GlobalContext) SqlContext() SqlContext {
	return g.sqlContext
}

func (g *GlobalContext) ConnNodeContext() *ConnNodeContext {
	return g.connNodeContext
}

func (g *GlobalContext) BroadcastContext() *BroadcastContext {
	return g.broadcastContext
}
