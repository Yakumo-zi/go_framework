package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Context struct {
	req    *http.Request
	resp   http.ResponseWriter
	params map[string]string
	query  url.Values
}

func (c *Context) PathParam(key string) StringValue {
	val, ok := c.params[key]
	if !ok {
		return StringValue{err: fmt.Errorf("path param %s not found", key), val: ""}
	}
	return StringValue{val: val, err: nil}
}

func (c *Context) QueryParam(key string) StringValue {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	val := c.query.Get(key)
	if val == "" {
		return StringValue{err: fmt.Errorf("query param %s not found", key), val: ""}
	}
	return StringValue{val: val, err: nil}
}
func (c *Context) FormParam(key string) StringValue {
	c.req.ParseForm()
	val := c.req.PostForm.Get(key)
	if val == "" {
		return StringValue{err: fmt.Errorf("form param %s not found", key), val: ""}
	}
	return StringValue{val: val, err: nil}
}
func (c *Context) BindJson(v any) error {
	return json.NewDecoder(c.req.Body).Decode(v)
}
func (c *Context) JSON(status int, v any) {
	c.resp.Header().Set("Content-Type", "application/json")
	c.resp.WriteHeader(status)
	_ = json.NewEncoder(c.resp).Encode(v)
}

type StringValue struct {
	val string
	err error
}

func (s StringValue) String() (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.val, nil
}
func (s StringValue) AsInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	return strconv.ParseInt(s.val, 10, 64)
}

func (s StringValue) BindJson(v any) error {
	return json.Unmarshal([]byte(s.val), v)
}
