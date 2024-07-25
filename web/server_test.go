package web

import (
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	s := &HTTPServer{
		addr: ":8080",
		router: router{
			trees: make(map[string]*node),
		},
	}
	s.Get("/hello", func(ctx *Context) {
		ctx.JSON(http.StatusOK, map[string]string{"hello": "world"})
	})
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	s.Post("/user", func(ctx *Context) {
		var user User
		if err := ctx.BindJson(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, user)
	})
	s.Get("/user/:id", func(ctx *Context) {
		id, err := ctx.PathParam("id").String()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, map[string]string{"id": id})
	})
	s.Start(s.addr)
}
