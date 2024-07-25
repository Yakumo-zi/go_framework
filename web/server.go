package web

import (
	"fmt"
	"net/http"
)

type HandleFunc func(ctx *Context)
type Server interface {
	http.Handler
	//AddRoute 将路由注册到路由树上
	//method HTTP方法
	//path 路由路径
	//handleFunc 业务逻辑
	addRoute(method string, path string, handleFunc HandleFunc)
	Start(addr string)
}

var _ Server = (*HTTPServer)(nil)

type HTTPServer struct {
	router
	addr string
}

// Get 基于AddRoute的衍生方法
func (h *HTTPServer) Get(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodGet, path, handleFunc)
}
func (h *HTTPServer) Post(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPost, path, handleFunc)
}
func (h *HTTPServer) Put(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPut, path, handleFunc)
}
func (h *HTTPServer) Delete(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodDelete, path, handleFunc)
}
func (h *HTTPServer) Patch(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPatch, path, handleFunc)
}
func (h *HTTPServer) Options(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodOptions, path, handleFunc)
}
func (h *HTTPServer) Head(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodHead, path, handleFunc)
}
func (h *HTTPServer) Connect(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodConnect, path, handleFunc)
}
func (h *HTTPServer) Trace(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodTrace, path, handleFunc)
}

func (h *HTTPServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ctx := &Context{
		req:    req,
		resp:   resp,
		params: make(map[string]string),
	}
	h.serve(ctx)
}

func (h *HTTPServer) serve(ctx *Context) {
	m, err := h.findNode(ctx.req.Method, ctx.req.URL.Path)
	if err != nil {
		ctx.resp.WriteHeader(http.StatusNotFound)
		return
	}
	ctx.params = m.params
	m.node.handleFunc(ctx)
}

func (h *HTTPServer) Start(addr string) {
	h.addr = addr
	fmt.Printf("Server started at http://%s\n", addr)
	_ = http.ListenAndServe(addr, h)
}
