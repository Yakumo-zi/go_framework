package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(conn),
	}
}

// Close implements Codec.
func (g *GobCodec) Close() error {
	return g.conn.Close()
}

// ReadBody implements Codec.
func (g *GobCodec) ReadBody(body any) error {
	return g.dec.Decode(body)
}

// ReadeHeader implements Codec.
func (g *GobCodec) ReadeHeader(h *Header) error {
	return g.dec.Decode(h)
}

// Write implements Codec.
func (g *GobCodec) Write(h *Header, body any) (err error) {
	defer func() {
		_ = g.buf.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()
	if err := g.enc.Encode(h); err != nil {
		log.Println("rpc codec:gob  error encoding header:", err)
		return err
	}
	if err := g.enc.Encode(body); err != nil {
		log.Println("rpc codec:gob  error encoding body:", err)
		return err
	}
	return nil
}
