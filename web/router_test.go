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
			assert.Panics(t, func() { r.addRoute(tc.method, tc.path, handleFunc) })
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

func TestFindNode(t *testing.T) {
	mock := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/users"},
		{"GET", "/users/:id"},
		{"DELETE", "/users/:id"},
		{"DELETE", "/users/:id/*"},
		{"PUT", "/goods/:id/:action"},
		{"PUT", "/users/*"},
		{"OPTION", "/users/(123)"},
		{"OPTION", "/users/(123)/*"},
	}
	testCase := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/users"},
		{"GET", "/users/123"},
		{"DELETE", "/users/123"},
		{"DELETE", "/users/123/"},
		{"DELETE", "/users/123/132131/gdggag"},
		{"PUT", "/users/123"},
		{"PUT", "/users/123/ststa/ggda/jjjj"},
		{"PUT", "/goods/123/action"},
		{"OPTION", "/users/123"},
		{"OPTION", "/users/123/321313gfag/fdafda"},
	}
	handleFunc := func(*Context) {}
	wanted := []*matchInfo{
		{
			node: &node{path: "/",
				handleFunc: handleFunc,
				children: map[string]*node{
					"users": {
						path:       "users",
						handleFunc: handleFunc,
						paramChild: &node{
							path:       ":id",
							handleFunc: handleFunc,
						},
					},
				},
			},
		},
		{
			node: &node{
				path:       "users",
				handleFunc: handleFunc,
				paramChild: &node{
					path:       ":id",
					handleFunc: handleFunc,
				},
			},
		},
		{
			node: &node{
				path:       ":id",
				handleFunc: handleFunc,
			},
			params: map[string]string{
				"id": "123",
			},
		},
		{
			node: &node{
				path:       ":id",
				handleFunc: handleFunc,
				starChild: &node{
					path:       "*",
					handleFunc: handleFunc,
				},
			},
			params: map[string]string{
				"id": "123",
			},
		},
		{
			node: &node{
				path:       "*",
				handleFunc: handleFunc,
			},
			params: map[string]string{
				"id": "123",
			},
		},
		{
			node: &node{
				path:       "*",
				handleFunc: handleFunc,
			},
			params: map[string]string{
				"id": "123",
			},
		},
		{
			node: &node{
				path:       "*",
				handleFunc: handleFunc,
			},
		},
		{
			node: &node{
				path:       "*",
				handleFunc: handleFunc,
			},
		},
		{
			node: &node{
				path:       ":action",
				handleFunc: handleFunc,
			},
			params: map[string]string{
				"id":     "123",
				"action": "action",
			},
		},
		{
			node: &node{
				path:       "(123)",
				handleFunc: handleFunc,
			},
		},
		{
			node: &node{
				path:       "*",
				handleFunc: handleFunc,
			},
		},
	}
	r := NewRouter()
	for _, tc := range mock {
		r.addRoute(tc.method, tc.path, handleFunc)
	}
	for i, tc := range testCase {
		matchInfo, err := r.findNode(tc.method, tc.path)
		if err != nil {
			t.Errorf("Test case %d failed", i)
			continue
		}
		if !matchInfo.equal(wanted[i]) {
			fmt.Printf("matchInfo: %+v,node:%+v\n", matchInfo, matchInfo.node)
			t.Errorf("Test case %d failed equal", i)
			continue
		}
		t.Logf("Test case %d passed", i)
	}
}
func (m *matchInfo) equal(y *matchInfo) bool {
	if !m.node.equal(y.node) {
		return false
	}
	if len(m.params) != len(y.params) {
		return false
	}
	for k, v := range m.params {
		if yv, ok := y.params[k]; !ok {
			return false
		} else {
			if v != yv {
				return false
			}
		}
	}
	return true
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
