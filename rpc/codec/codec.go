package codec

import "io"

type Codec interface {
	io.Closer
	ReadeHeader(*Header) error
	ReadBody(any) error
	Write(*Header, any) error
}

type Header struct {
	ServiceMethod string //format "Service.Method"
	Seq           uint64 //sequence number chosen client
	Error         string
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
