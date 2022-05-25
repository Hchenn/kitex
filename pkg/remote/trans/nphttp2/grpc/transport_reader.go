package grpc

import (
	"github.com/cloudwego/netpoll"
)

var _ netpoll.Reader = &recvBufferReader{}

func (r *recvBufferReader) Next(n int) (p []byte, err error) {
	err = r.check(n)
	if err != nil {
		return nil, err
	}
	return r.last.Next(n)
}

func (r *recvBufferReader) Peek(n int) (buf []byte, err error) {
	err = r.check(n)
	if err != nil {
		return nil, err
	}
	return r.last.Peek(n)
}

func (r *recvBufferReader) Skip(n int) (err error) {
	err = r.check(n)
	if err != nil {
		return err
	}
	return r.last.Skip(n)
}

func (r *recvBufferReader) Until(delim byte) (line []byte, err error) {
	panic("until is not implement")
}

func (r *recvBufferReader) ReadString(n int) (s string, err error) {
	err = r.check(n)
	if err != nil {
		return s, err
	}
	return r.last.ReadString(n)
}

func (r *recvBufferReader) ReadBinary(n int) (p []byte, err error) {
	err = r.check(n)
	if err != nil {
		return p, err
	}
	return r.last.ReadBinary(n)
}

func (r *recvBufferReader) ReadByte() (b byte, err error) {
	err = r.check(1)
	if err != nil {
		return b, err
	}
	return r.last.ReadByte()
}

func (r *recvBufferReader) Slice(n int) (_ netpoll.Reader, err error) {
	err = r.check(n)
	if err != nil {
		return nil, err
	}
	return r.last.Slice(n)
}

func (r *recvBufferReader) Release() (err error) {
	return r.last.Release()
}

func (r *recvBufferReader) Len() (length int) {
	return r.last.Len()
}

func (r *recvBufferReader) check(n int) (err error) {
	if r.err != nil {
		return r.err
	}
	for r.last.Len() < n {
		if r.closeStream != nil {
			r.err = r.fillClient()
		} else {
			r.err = r.fill()
		}
		if r.err != nil {
			return r.err
		}
	}
	return nil
}

var _ netpoll.Reader = &transportReader{}

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

var _ netpoll.Reader = &Stream{}

func (s *Stream) Next(n int) (p []byte, err error) {
	// Don't request a read if there was an error earlier
	if er := s.trReader.(*transportReader).er; er != nil {
		return nil, er
	}
	s.requestRead(n)
	return s.trReader.Next(n)
}

func (s *Stream) Peek(n int) (buf []byte, err error) {
	// Don't request a read if there was an error earlier
	if er := s.trReader.(*transportReader).er; er != nil {
		return nil, er
	}
	s.requestRead(n)
	return s.trReader.Peek(n)
}

func (s *Stream) Skip(n int) (err error) {
	// Don't request a read if there was an error earlier
	if er := s.trReader.(*transportReader).er; er != nil {
		return er
	}
	s.requestRead(n)
	return s.trReader.Skip(n)
}

func (s *Stream) Until(delim byte) (line []byte, err error) {
	panic("until is not implement")
}

func (s *Stream) ReadString(n int) (_ string, err error) {
	// Don't request a read if there was an error earlier
	if er := s.trReader.(*transportReader).er; er != nil {
		return "", er
	}
	s.requestRead(n)
	return s.trReader.ReadString(n)
}

func (s *Stream) ReadBinary(n int) (p []byte, err error) {
	// Don't request a read if there was an error earlier
	if er := s.trReader.(*transportReader).er; er != nil {
		return p, er
	}
	s.requestRead(n)
	return s.trReader.ReadBinary(n)
}

func (s *Stream) ReadByte() (b byte, err error) {
	// Don't request a read if there was an error earlier
	if er := s.trReader.(*transportReader).er; er != nil {
		return 0, er
	}
	s.requestRead(1)
	return s.trReader.ReadByte()
}

func (s *Stream) Slice(n int) (r netpoll.Reader, err error) {
	// Don't request a read if there was an error earlier
	if er := s.trReader.(*transportReader).er; er != nil {
		return nil, er
	}
	s.requestRead(n)
	return s.trReader.Slice(n)
}

func (s *Stream) Release() (err error) {
	return s.trReader.Release()
}

func (s *Stream) Len() (length int) {
	return s.trReader.Len()
}
