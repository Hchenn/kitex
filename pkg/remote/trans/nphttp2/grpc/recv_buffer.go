package grpc

import (
	"context"
	"sync"

	"github.com/cloudwego/netpoll"
)

// recvMsg represents the received msg from the transport. All transport
// protocol specific info has been removed.
type recvMsg struct {
	buffer *netpoll.LinkBuffer

	// nil: received some data
	// io.EOF: stream is completed. data is nil.
	// other non-nil error: transport failure. data is nil.
	err error
}

// recvBuffer is an unbounded channel of recvMsg structs.
//
// Note: recvBuffer differs from buffer.Unbounded only in the fact that it
// holds a channel of recvMsg structs instead of objects implementing "item"
// interface. recvBuffer is written to much more often and using strict recvMsg
// structs helps avoid allocation in "recvBuffer.put"
type recvBuffer struct {
	c       chan recvMsg
	mu      sync.Mutex
	backlog []recvMsg
	err     error
}

func newRecvBuffer() *recvBuffer {
	b := &recvBuffer{
		c: make(chan recvMsg, 1),
	}
	return b
}

func (b *recvBuffer) put(r recvMsg) {
	b.mu.Lock()
	if b.err != nil {
		b.mu.Unlock()
		// An error had occurred earlier, don't accept more
		// data or errors.
		return
	}
	b.err = r.err
	if len(b.backlog) == 0 {
		select {
		case b.c <- r:
			b.mu.Unlock()
			return
		default:
		}
	}
	b.backlog = append(b.backlog, r)
	b.mu.Unlock()
}

func (b *recvBuffer) load() {
	b.mu.Lock()
	if len(b.backlog) > 0 {
		select {
		case b.c <- b.backlog[0]:
			b.backlog[0] = recvMsg{}
			b.backlog = b.backlog[1:]
		default:
		}
	}
	b.mu.Unlock()
}

// get returns the channel that receives a recvMsg in the buffer.
//
// Upon receipt of a recvMsg, the caller should call load to send another
// recvMsg onto the channel if there is any.
func (b *recvBuffer) get() <-chan recvMsg {
	return b.c
}

var _ netpoll.Reader = &recvBufferReader{}

// recvBufferReader implements netpoll.Reader interface to read the data from
// recvBuffer.
type recvBufferReader struct {
	closeStream func(error) // Closes the client transport stream with the given error and nil trailer metadata.
	ctx         context.Context
	ctxDone     <-chan struct{} // cache of ctx.Done() (for performance).
	recv        *recvBuffer
	last        *netpoll.LinkBuffer // Stores the remaining data in the previous calls.
	err         error
}

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
	if r.last == nil {
		return nil
	}
	return r.last.Release()
}

func (r *recvBufferReader) Len() (length int) {
	if r.last == nil {
		return 0
	}
	return r.last.Len()
}

func (r *recvBufferReader) check(n int) (err error) {
	if r.err != nil {
		return r.err
	}
	for r.last == nil || r.last.Len() < n {
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

func (r *recvBufferReader) fill() (err error) {
	select {
	case <-r.ctxDone:
		return ContextErr(r.ctx.Err())
	case m := <-r.recv.get():
		return r.fillAdditional(m)
	}
}

func (r *recvBufferReader) fillClient() (err error) {
	// If the context is canceled, then closes the stream with nil metadata.
	// closeStream writes its error parameter to r.recv as a recvMsg.
	// r.readAdditional acts on that message and returns the necessary error.
	select {
	case <-r.ctxDone:
		// Note that this adds the ctx error to the end of recv buffer, and
		// reads from the head. This will delay the error until recv buffer is
		// empty, thus will delay ctx cancellation in Recv().
		//
		// It's done this way to fix a race between ctx cancel and trailer. The
		// race was, stream.Recv() may return ctx error if ctxDone wins the
		// race, but stream.Trailer() may return a non-nil md because the stream
		// was not marked as done when trailer is received. This closeStream
		// call will mark stream as done, thus fix the race.
		//
		// TODO: delaying ctx error seems like a unnecessary side effect. What
		// we really want is to mark the stream as done, and return ctx error
		// faster.
		r.closeStream(ContextErr(r.ctx.Err()))
		m := <-r.recv.get()
		return r.fillAdditional(m)
	case m := <-r.recv.get():
		return r.fillAdditional(m)
	}
}

func (r *recvBufferReader) fillAdditional(m recvMsg) (err error) {
	r.recv.load()
	if m.err != nil {
		return m.err
	}
	if r.last == nil {
		r.last = m.buffer
		return nil
	}
	return r.last.Append(m.buffer)
}
