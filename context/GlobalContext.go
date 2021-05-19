package context

import "tursom-im/config"

type GlobalContext struct {
	tokenContext    *TokenContext
	attrContext     *AttrContext
	userConnContext *UserConnContext
	msgIdContext    *MsgIdContext
	cfg             *config.Config
	sqlContext      SqlContext
}

func NewGlobalContext(config *config.Config) *GlobalContext {
	g := &GlobalContext{
		tokenContext:    NewTokenContext(),
		attrContext:     NewAttrContext(),
		userConnContext: NewUserConnContext(),
		msgIdContext:    NewMsgIdContext(),
		cfg:             config,
		sqlContext:      NewSqliteSqlContext(),
	}
	g.userConnContext.Init(g)
	g.sqlContext.Init(g)
	g.tokenContext.Init(g)
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
