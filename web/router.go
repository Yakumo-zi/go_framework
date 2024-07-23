package web

import (
	"strings"
)

type node struct {
	path       string
	children   map[string]*node
	starChild  *node
	regexChild *node
	paramChild *node
	handleFunc HandleFunc
}

type matchInfo struct {
	params    map[string]string
	hanleFunc HandleFunc
}

type router struct {
	trees map[string]*node
}

func NewRouter() *router {
	return &router{
		trees: make(map[string]*node),
	}
}

func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {
	root, ok := r.trees[method]
	if !ok {
		root = &node{
			children: make(map[string]*node),
		}
		root.path = "/"
		r.trees[method] = root
	}
	if path == "/" {
		root.handleFunc = handleFunc
		return
	}
	segs := strings.Split(path[1:], "/")
	for _, seg := range segs {
		if seg == " " {
			panic("web:不允许连续的 '/' ")
		}
		ret := childOrCreate(root, seg)
		root = ret
	}
	root.handleFunc = handleFunc
}

func childOrCreate(root *node, seg string) *node {
	ret, ok := root.children[seg]
	if ok {
		return ret
	}
	ret = &node{
		path:     seg,
		children: make(map[string]*node),
	}
	if seg == "*" {
		if root.paramChild != nil {
			panic("web:不允许同时注册参数路径与通配符路径")
		}
		if root.starChild != nil {
			return root.starChild
		} else {
			root.starChild = ret
			return ret
		}
	}
	if seg[0:1] == ":" {
		if root.starChild != nil {
			panic("web:不允许同时注册参数路劲和通配符路径")
		}
		if root.paramChild != nil {
			return root.paramChild
		} else {
			root.paramChild = ret
			return ret
		}
	}
	if seg[0:1] == "(" {
		if root.regexChild != nil {
			return root.regexChild
		} else {
			root.regexChild = ret
			return ret
		}
	}
	root.children[seg] = ret
	return ret
}
