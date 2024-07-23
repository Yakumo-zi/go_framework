package web

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	seg := ":id"
	fmt.Println(seg[0:1] == ":")

}

func TestAddRoute(t *testing.T) {
	testCases := []struct {
		method string
		path   string
		panic  bool
	}{
		{"GET", "/", false},
		{"GET", "/users", false},
		{"GET", "/users/:id", false},
		{"DELETE", "/users/:id", false},
		{"PUT", "/users/*", false},
		{"DELETE", "/users/:id/*", false},
		{"GET", "///", true},
		{"GET", "/a///b//", true},
	}

	handleFunc := func(*Context) {}
	wanted := []*router{
		{
			trees: map[string]*node{
				"GET": {
					path:       "/",
					children:   make(map[string]*node),
					handleFunc: handleFunc,
				},
			},
		},
		{
			trees: map[string]*node{
				"GET": {
					path: "/",
					children: map[string]*node{
						"users": {
							path:       "users",
							children:   make(map[string]*node),
							handleFunc: handleFunc,
						},
					},
				},
			},
		},
		{
			trees: map[string]*node{
				"GET": {
					path: "/",
					children: map[string]*node{
						"users": {
							path:     "users",
							children: make(map[string]*node),
							paramChild: &node{
								path:       ":id",
								children:   make(map[string]*node),
								handleFunc: handleFunc,
							},
						},
					},
				},
			},
		},
		{
			trees: map[string]*node{
				"DELETE": {
					path: "/",
					children: map[string]*node{
						"users": {
							path:     "users",
							children: make(map[string]*node),
							paramChild: &node{
								path:       ":id",
								children:   make(map[string]*node),
								handleFunc: handleFunc,
							},
						},
					},
				},
			},
		},
		{
			trees: map[string]*node{
				"PUT": {
					path: "/",
					children: map[string]*node{
						"users": {
							path:     "users",
							children: make(map[string]*node),
							starChild: &node{
								path:       "*",
								children:   make(map[string]*node),
								handleFunc: handleFunc,
							},
						},
					},
				},
			},
		},
		{
			trees: map[string]*node{
				"DELETE": {
					path: "/",
					children: map[string]*node{
						"users": {
							path:     "users",
							children: make(map[string]*node),
							paramChild: &node{
								path:     ":id",
								children: make(map[string]*node),
								starChild: &node{
									path:       "*",
									children:   make(map[string]*node),
									handleFunc: handleFunc,
								},
							},
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		r := NewRouter()
		if !tc.panic {
			r.addRoute(tc.method, tc.path, handleFunc)
			if !r.equal(wanted[i]) {
				t.Errorf("Test case %d failed", i)
				r.trees[tc.method].print("")
				wanted[i].trees[tc.method].print("")
			}
		} else {
			assert.Panics(t, func() { r.addRoute(tc.method, tc.path, handleFunc) }, "Panic")
		}
		t.Logf("Test case %d passed", i)
	}
}

func TestAddRoutes(t *testing.T) {
	testCases := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/users"},
		{"GET", "/users/:id"},
		{"DELETE", "/users/:id"},
		{"PUT", "/users/*"},
		{"DELETE", "/users/:id/*"},
	}

	handleFunc := func(*Context) {}
	wanted := &router{
		trees: map[string]*node{
			"GET": {
				path:       "/",
				handleFunc: handleFunc,
				children: map[string]*node{
					"users": {
						path:       "users",
						handleFunc: handleFunc,
						paramChild: &node{
							path:       ":id",
							children:   make(map[string]*node),
							handleFunc: handleFunc,
						},
					},
				},
			},
			"DELETE": {
				path: "/",
				children: map[string]*node{
					"users": {
						path: "users",
						paramChild: &node{
							path:       ":id",
							handleFunc: handleFunc,
							children:   make(map[string]*node),
							starChild: &node{
								path:       "*",
								children:   make(map[string]*node),
								handleFunc: handleFunc,
							},
						},
					},
				},
			},
			"PUT": {
				path: "/",
				children: map[string]*node{
					"users": {
						path: "users",
						starChild: &node{
							path:       "*",
							children:   make(map[string]*node),
							handleFunc: handleFunc,
						},
					},
				},
			},
		},
	}
	r := NewRouter()
	for _, tc := range testCases {
		r.addRoute(tc.method, tc.path, handleFunc)
	}
	if !r.equal(wanted) {
		t.Errorf("Test case failed")
	} else {
		t.Logf("Test case passed")
	}
}

func (r *router) equal(y *router) bool {
	if len(r.trees) != len(y.trees) {
		return false
	}
	for k, v := range r.trees {
		if yv, ok := y.trees[k]; !ok {
			return false
		} else {
			if !v.equal(yv) {
				return false
			}
		}
	}
	return true
}
func (n *node) print(prefix string) {
	fmt.Println(prefix+"|", n.path)
	for _, v := range n.children {
		v.print(prefix + "  ")
	}
	if n.starChild != nil {
		n.starChild.print(prefix + "  *")
	}
	if n.paramChild != nil {
		n.paramChild.print(prefix + "  p")
	}
}

func (n *node) equal(y *node) bool {
	if n.path != y.path {
		return false
	}
	nHandleFunc := reflect.ValueOf(n.handleFunc)
	yHandleFunc := reflect.ValueOf(y.handleFunc)
	if nHandleFunc != yHandleFunc {
		return false
	}

	if len(n.children) != len(y.children) {
		return false
	}
	if y.starChild != nil {
		return y.starChild.equal(n.starChild)
	}
	if y.paramChild != nil {
		return y.paramChild.equal(n.paramChild)
	}
	for k, v := range n.children {
		if yv, ok := y.children[k]; !ok {
			return false
		} else {
			if !v.equal(yv) {
				return false
			}
		}
	}
	return true
}
