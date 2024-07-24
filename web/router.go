package web

import (
	"fmt"
	"regexp"
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

func (r *router) findNode(method, path string) (*matchInfo, error) {
	root, ok := r.trees[method]
	if !ok {
		return nil, fmt.Errorf("web:请求路径 %s 未注册", method+" "+path)
	}
	if path == "/" {
		return &matchInfo{
			hanleFunc: root.handleFunc,
		}, nil
	}
	matchInfo := &matchInfo{
		params: make(map[string]string),
	}
	segs := strings.Split(path[1:], "/")
	for _, seg := range segs {
		ret, ok := root.children[seg]
		if !ok {
			if root.starChild != nil {
				root = root.starChild
				continue
			}
			if root.paramChild != nil {
				matchInfo.params[root.paramChild.path[1:]] = seg
				root = root.paramChild
				continue
			}
			if root.regexChild != nil {
				matched, err := regexp.Match(root.regexChild.path, []byte(seg))
				if err != nil {
					return nil, err
				}
				if !matched {
					return nil, fmt.Errorf("web:请求路径 %s 未注册", method+" "+path)
				}
				root = root.regexChild
				continue
			}
			if root.path == "*" {
				continue
			}
			return nil, fmt.Errorf("web:请求路径 %s 未注册", method+" "+path)
		}
		root = ret
	}
	matchInfo.hanleFunc = root.handleFunc
	return matchInfo, nil
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
		if root.regexChild != nil {
			panic("web:不允许同时注册正则路径和通配符路径")
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
			panic("web:不允许同时注册参数路路径和通配符路径")
		}
		if root.regexChild != nil {
			panic("web:不允许同时注册正则路径和参数路径")
		}
		if root.paramChild != nil {
			return root.paramChild
		} else {
			root.paramChild = ret
			return ret
		}
	}
	if seg[0:1] == "(" {
		if root.starChild != nil {
			panic("web:不允许同时注册正则路径和通配符路径")
		}
		if root.paramChild != nil {
			panic("web:不允许同时注册正则路径和参数路径")
		}
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
