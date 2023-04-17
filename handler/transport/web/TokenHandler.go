package web

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"

	"github.com/tursom-im/context"
)

type TokenHandler struct {
	lang.BaseObject
	globalContext *context.GlobalContext
}

func NewTokenHandler(ctx *context.GlobalContext) *TokenHandler {
	return &TokenHandler{
		globalContext: ctx,
	}
}

func (h *TokenHandler) InitWebHandler(router Router) {
	router.POST("/token", h.FlushToken)
	router.PUT("/user", h.NewUser)
}

func (h *TokenHandler) FlushToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	exceptions.CheckNil(h)
	var err error = nil
	defer handleError(w, err)

	appId := h.globalContext.Config().Admin.CheckAdmin(r)
	if appId == nil {
		w.WriteHeader(502)
		return
	}
	query := r.URL.Query()
	uid := query["uid"]
	if len(uid) == 0 {
		w.WriteHeader(400)
		return
	}
	token, err := h.globalContext.Token().FlushToken(uid[0])
	if err != nil {
		err = exceptions.Package(err)
		exceptions.Print(err)
		return
	}
	if len(token) == 0 {
		w.WriteHeader(500)
		return
	}

	_, err = w.Write([]byte(token))
	if err != nil {
		err = exceptions.Package(err)
		exceptions.Print(err)
		return
	}
}

func (h *TokenHandler) NewUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	exceptions.CheckNil(h)
	var err error = nil
	defer handleError(w, err)

	appId := h.globalContext.Config().Admin.CheckAdmin(r)
	if appId == nil {
		w.WriteHeader(502)
		return
	}

	user, err := h.globalContext.Sql().GetUserTableContext().CreateUser()
	if err != nil {
		err = exceptions.Package(err)
		exceptions.Print(err)
		return
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		err = exceptions.Package(err)
		exceptions.Print(err)
		return
	}

	_, err = w.Write(userBytes)
	if err != nil {
		err = exceptions.Package(err)
		exceptions.Print(err)
		return
	}
}

func handleError(w http.ResponseWriter, err error) {
	if err != nil {
		exceptions.Print(err)
		w.WriteHeader(500)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			exceptions.Print(err)
		}
	}
}
