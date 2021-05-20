package handler

import "github.com/julienschmidt/httprouter"

type WebHandler interface {
	InitWebHandler(basePath string, router *httprouter.Router)
}
