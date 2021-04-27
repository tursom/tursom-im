package context

type GlobalContext struct {
	tokenContext TokenContext
	attrContext  AttrContext
}

func (g *GlobalContext) AttrContext() AttrContext {
	return g.attrContext
}

func (g *GlobalContext) TokenContext() TokenContext {
	return g.tokenContext
}
