package grpc

import (
	"github.com/cloudwego/netpoll"
)

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
