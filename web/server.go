package web

import "net/http"

type HandleFunc func(ctx *Context)
type Server interface {
	http.Handler
	//AddRoute 将路由注册到路由树上
	//method HTTP方法
	//path 路由路径
	//handleFunc 业务逻辑
	AddRoute(method string, path string, handleFunc HandleFunc)
	Start(addr string)
}

var _ Server = (*HTTPServer)(nil)

type HTTPServer struct {
	addr string
}

func (h *HTTPServer) AddRoute(method string, path string, handleFunc HandleFunc) {
	panic("unimplemented")
}

// Get 基于AddRoute的衍生方法
func (h *HTTPServer) Get(path string, handleFunc HandleFunc) {
	h.AddRoute(http.MethodGet, "", handleFunc)
}

func (h *HTTPServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ctx := &Context{
		req:  req,
		resp: resp,
	}
	h.serve(ctx)
	panic("unimplemented")
}

func (h *HTTPServer) serve(ctx *Context) {

}

func (h *HTTPServer) Start(addr string) {
	panic("unimplemented")
}
