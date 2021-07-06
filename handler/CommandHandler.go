package handler

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"tursom-im/context"
)

type CommandHandler struct {
	globalContext context.GlobalContext
}

func NewCommandHandler(globalContext context.GlobalContext) *CommandHandler {
	return &CommandHandler{globalContext: globalContext}
}

func (c *CommandHandler) InitWebHandler(basePath string, router *httprouter.Router) {
	router.GET(basePath+"/close", c.CloseServer)
}

func (t *CommandHandler) CloseServer(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error = nil
	defer handleError(w, err)

	appId := t.globalContext.Config().Admin.CheckAdmin(r)
	if appId == nil {
		w.WriteHeader(502)
		return
	}

	os.Exit(0)
}
