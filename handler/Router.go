package handler

import (
	"github.com/julienschmidt/httprouter"
)

type (
	Router interface {
		GET(path string, handle httprouter.Handle)
		HEAD(path string, handle httprouter.Handle)
		OPTIONS(path string, handle httprouter.Handle)
		POST(path string, handle httprouter.Handle)
		PUT(path string, handle httprouter.Handle)
		PATCH(path string, handle httprouter.Handle)
		DELETE(path string, handle httprouter.Handle)
	}
	SubRouter struct {
		router Router
		path   string
	}
)

func (r *SubRouter) GET(path string, handle httprouter.Handle) {
	r.router.GET(r.path+path, handle)
}

func (r *SubRouter) HEAD(path string, handle httprouter.Handle) {
	r.router.HEAD(r.path+path, handle)
}

func (r *SubRouter) OPTIONS(path string, handle httprouter.Handle) {
	r.router.OPTIONS(r.path+path, handle)
}

func (r *SubRouter) POST(path string, handle httprouter.Handle) {
	r.router.POST(r.path+path, handle)
}

func (r *SubRouter) PUT(path string, handle httprouter.Handle) {
	r.router.PUT(r.path+path, handle)
}

func (r *SubRouter) PATCH(path string, handle httprouter.Handle) {
	r.router.PATCH(r.path+path, handle)
}

func (r *SubRouter) DELETE(path string, handle httprouter.Handle) {
	r.router.DELETE(r.path+path, handle)
}
