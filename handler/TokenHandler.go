package handler

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/tursom/GoCollections/exceptions"
	"net/http"
	"tursom-im/context"
)

type TokenHandler struct {
	globalContext *context.GlobalContext
}

func NewTokenHandler(ctx *context.GlobalContext) *TokenHandler {
	return &TokenHandler{
		globalContext: ctx,
	}
}

func (t *TokenHandler) InitWebHandler(basePath string, router *httprouter.Router) {
	router.POST(basePath+"/token", t.FlushToken)
	router.PUT(basePath+"/user", t.NewUser)
}

func (t *TokenHandler) FlushToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error = nil
	defer handleError(w, err)

	appId := t.globalContext.Config().Admin.CheckAdmin(r)
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
	token, err := t.globalContext.TokenContext().FlushToken(uid[0])
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

func (t *TokenHandler) NewUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var err error = nil
	defer handleError(w, err)

	appId := t.globalContext.Config().Admin.CheckAdmin(r)
	if appId == nil {
		w.WriteHeader(502)
		return
	}

	user, err := t.globalContext.SqlContext().GetUserTableContext().CreateUser()
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
