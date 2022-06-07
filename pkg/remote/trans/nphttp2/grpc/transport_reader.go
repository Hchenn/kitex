package grpc

import (
	"github.com/cloudwego/netpoll"
)

var _ netpoll.Reader = &recvBufferReader{}

func (r *recvBufferReader) Next(n int) (p []byte, err error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.last.Len() == 0 {
		// fill
		if r.closeStream != nil {
			r.err = r.fillClient()
		} else {
			r.err = r.fill()
		}
		if r.err != nil {
			return p, r.err
		}
	}

	if r.last.Len() >= n {
		return r.last.Next(n)
	}

	p = make([]byte, n)
	pIdx, pDelta := 0, n
	for {
		l := r.last.Len()
		switch {
		case l >= pDelta:
			b1, _ := r.last.Next(pDelta)
			copy(p[pIdx:], b1)
			return p, nil
		case l > 0:
			b1, _ := r.last.Next(l)
			pIdx += copy(p[pIdx:], b1)
			pDelta = n - pIdx
		}

		// fill
		if r.closeStream != nil {
			r.err = r.fillClient()
		} else {
			r.err = r.fill()
		}
		if r.err != nil {
			return p, r.err
		}
	}
}

func (r *recvBufferReader) Peek(n int) (buf []byte, err error) {
	panic("peek is not implement")
}

func (r *recvBufferReader) Skip(n int) (err error) {
	panic("skip is not implement")
}

func (r *recvBufferReader) Until(delim byte) (line []byte, err error) {
	panic("until is not implement")
}

func (r *recvBufferReader) ReadString(n int) (s string, err error) {
	panic("readstring is not implement")
}

func (r *recvBufferReader) ReadBinary(n int) (p []byte, err error) {
	panic("readbinary is not implement")
}

func (r *recvBufferReader) ReadByte() (b byte, err error) {
	panic("readbyte is not implement")
}

func (r *recvBufferReader) Slice(n int) (_ netpoll.Reader, err error) {
	panic("slice is not implement")
}

func (r *recvBufferReader) Release() (err error) {
	return r.last.Release()
}

func (r *recvBufferReader) Len() (length int) {
	return r.last.Len()
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
