package grpc

import (
	"github.com/cloudwego/netpoll"
)

var _ netpoll.Reader = &transportReader{}

// transportReader reads all the data available for this Stream from the transport and
// passes them into the decoder, which converts them into a gRPC message stream.
// The error is io.EOF when the stream is done or another non-nil error if
// the stream broke.
type transportReader struct {
	reader netpoll.Reader
	// The handler to control the window update procedure for both this
	// particular stream and the associated transport.
	windowHandler func(int)
	er            error
}

func (t *transportReader) Next(n int) (p []byte, err error) {
	p, err = t.reader.Next(n)
	if err != nil {
		t.er = err
		return
	}
	t.windowHandler(n)
	return
}

func (t *transportReader) Peek(n int) (buf []byte, err error) {
	buf, err = t.reader.Peek(n)
	if err != nil {
		t.er = err
		return
	}
	t.windowHandler(n)
	return
}

func (t *transportReader) Skip(n int) (err error) {
	err = t.reader.Skip(n)
	if err != nil {
		t.er = err
		return
	}
	t.windowHandler(n)
	return
}

func (t *transportReader) Until(delim byte) (line []byte, err error) {
	panic("until is not implement")
}

func (t *transportReader) ReadString(n int) (s string, err error) {
	s, err = t.reader.ReadString(n)
	if err != nil {
		t.er = err
		return
	}
	t.windowHandler(n)
	return
}

func (t *transportReader) ReadBinary(n int) (p []byte, err error) {
	p, err = t.reader.ReadBinary(n)
	if err != nil {
		t.er = err
		return
	}
	t.windowHandler(n)
	return
}

func (t *transportReader) ReadByte() (b byte, err error) {
	b, err = t.reader.ReadByte()
	if err != nil {
		t.er = err
		return
	}
	t.windowHandler(1)
	return
}

func (t *transportReader) Slice(n int) (r netpoll.Reader, err error) {
	r, err = t.reader.Slice(n)
	if err != nil {
		t.er = err
		return
	}
	t.windowHandler(n)
	return
}

func (t *transportReader) Release() (err error) {
	return t.reader.Release()
}

func (t *transportReader) Len() (length int) {
	return t.reader.Len()
}
