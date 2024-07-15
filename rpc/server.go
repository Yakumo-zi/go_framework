package rpc

import (
	"encoding/json"
	"errors"
	"go/ast"
	"io"
	"log"
	"net"
	"reflect"
	"rpc/codec"
	"strings"
	"sync"
	"sync/atomic"
)

const MagicNumber = 0x3bef5c

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}
var invalidRequest = struct{}{}

type request struct {
	h           *codec.Header
	argv, reply reflect.Value
	mtype       *methodType
	svc         *service
}

type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint64
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}
func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *methodType) newReply() reflect.Value {
	replyV := reflect.New(m.ReplyType.Elem())

	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyV.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyV.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))

	}
	return replyV
}

type service struct {
	name   string
	typ    reflect.Type
	rcvr   reflect.Value
	method map[string]*methodType
}

func newService(rcvr any) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()

	s.typ = reflect.TypeOf(rcvr)
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server:%s is not a valid service name", s.name)
	}
	s.registerMethods()
	return s
}

func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}

		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server:register %s.%s\n", s.name, method.Name)
	}
}
func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

type Server struct {
	serviceMap sync.Map
}

func (server *Server) Register(rcvr any) error {
	s := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc: service already defined:" + s.name)
	}
	return nil
}

func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server:service/method request ill-formed:" + serviceMethod)
		return
	}

	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server:can't find service:" + serviceName)
		return
	}
	svc = svci.(*service)
	mtype, ok = svc.method[methodName]
	if !ok {
		err = errors.New("rpc server:can't find method:" + methodName)
		return
	}
	return svc, mtype, nil
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (s *Server) ServeConn(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: option error:", err)
		return
	}

	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server:invalid magic number %x\n", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server:invalid codec type %s\n", opt.CodecType)
		return
	}

	s.serveCodec(f(conn))
}

func (s *Server) serveCodec(codec codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := s.readRequest(codec)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			s.sendResponse(codec, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(codec, req, sending, wg)
	}
}

func (s *Server) readRequest(cc codec.Codec) (req *request, err error) {
	var h codec.Header

	if err = cc.ReadeHeader(&h); err != nil {
		if err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
			log.Println("rpc server: read head error:", err)
		}
		return nil, err
	}
	req = &request{h: &h}
	req.svc, req.mtype, err = s.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}

	req.argv = req.mtype.newArgv()
	req.reply = req.mtype.newReply()

	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body error:", err)
		return req, err
	}
	return req, err

}
func (s *Server) sendResponse(cc codec.Codec, header *codec.Header, body any, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()
	if err := cc.Write(header, body); err != nil {
		log.Println("rpc server:write response error:", err)
	}

}
func (s *Server) handleRequest(cc codec.Codec, req *request, mutex *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	err := req.svc.call(req.mtype, req.argv, req.reply)
	if err != nil {
		req.h.Error = err.Error()
		s.sendResponse(cc, req.h, invalidRequest, mutex)
	}
	s.sendResponse(cc, req.h, req.reply.Interface(), mutex)

}

func (s *Server) Accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go s.ServeConn(conn)

	}
}
func Accept(listener net.Listener) {
	DefaultServer.Accept(listener)
}

func Register(rcvr any) error { return DefaultServer.Register(rcvr) }
