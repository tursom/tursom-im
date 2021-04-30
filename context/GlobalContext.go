package context

import "tursom-im/config"

type GlobalContext struct {
	tokenContext    *TokenContext
	attrContext     *AttrContext
	userConnContext *UserConnContext
	cfg             *config.Config
}

func NewGlobalContext(config *config.Config) *GlobalContext {
	return &GlobalContext{
		tokenContext:    NewTokenContext(),
		attrContext:     NewAttrContext(),
		userConnContext: NewUserConnContext(),
		cfg:             config,
	}
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
