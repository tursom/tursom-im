package handler

import (
	"github.com/julienschmidt/httprouter"
	"github.com/tursom-im/context"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"net/http"
	"os"
)

type CommandHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func NewCommandHandler(globalContext *context.GlobalContext) *CommandHandler {
	return &CommandHandler{globalContext: globalContext}
}

func (h *CommandHandler) InitWebHandler(basePath string, router *httprouter.Router) {
	router.GET(basePath+"/close", h.CloseServer)
}

func (h *CommandHandler) CloseServer(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	exceptions.CheckNil(h)
	var err error = nil
	defer handleError(w, err)

	appId := h.globalContext.Config().Admin.CheckAdmin(r)
	if appId == nil {
		w.WriteHeader(502)
		return
	}

	os.Exit(0)
}
