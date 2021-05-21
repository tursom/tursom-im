package context

import (
	"tursom-im/config"
)

type GlobalContext struct {
	tokenContext    *TokenContext
	attrContext     *AttrContext
	userConnContext *UserConnContext
	msgIdContext    *MsgIdContext
	cfg             *config.Config
	sqlContext      SqlContext
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

	g := &GlobalContext{
		tokenContext:    tokenContext,
		attrContext:     attrContext,
		userConnContext: userConnContext,
		msgIdContext:    msgIdContext,
		cfg:             config,
		sqlContext:      sqlContext,
	}
	g.tokenContext.Init(g)
	g.userConnContext.Init(g)
	g.sqlContext.Init(g)
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
