package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"rpc/codec"
	"sync"
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
}

type Option struct {
	MagicNumber int
	CodecType   codec.Type
}
type Server struct{}

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
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read body error:", err)
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
	log.Println(req.h, req.argv.Elem())
	req.reply = reflect.ValueOf(fmt.Sprintf("rpc resp %d", req.h.Seq))
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
