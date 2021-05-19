package handler

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"tursom-im/context"
)

type TokenHandler struct {
	globalContext context.GlobalContext
}

func NewTokenHandler(ctx context.GlobalContext) *TokenHandler {
	return &TokenHandler{
		globalContext: ctx,
	}
}

func (t *TokenHandler) NewToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	appId := r.Header["AppId"][0]
	appToken := r.Header["AppToken"][0]
	if appId != t.globalContext.Config().Admin.Id || appToken != t.globalContext.Config().Admin.Password {
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
	if err != nil || len(token) == 0 {
		w.WriteHeader(500)
		return
	}
	_, err = w.Write([]byte(token))
	if err != nil {
		w.WriteHeader(500)
		return
	}
}
