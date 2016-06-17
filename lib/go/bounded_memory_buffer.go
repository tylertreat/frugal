package frugal

import "git.apache.org/thrift.git/lib/go/thrift"

// FBoundedMemoryBuffer implements TTransport using a bounded memory buffer.
// Writes which cause the buffer to exceed its size return ErrTooLarge.
type FBoundedMemoryBuffer struct {
	limit int
	*thrift.TMemoryBuffer
}

// NewFBoundedMemoryBuffer returns a new FBoundedMemoryBuffer with the given
// size limit.
func NewFBoundedMemoryBuffer(size int) *FBoundedMemoryBuffer {
	return &FBoundedMemoryBuffer{size, thrift.NewTMemoryBuffer()}
}

// Write the data to the buffer. Returns ErrTooLarge if the write would cause
// the buffer to exceed its limit.
func (f *FBoundedMemoryBuffer) Write(buf []byte) (int, error) {
	if len(buf)+f.Len() > f.limit {
		f.Reset()
		return 0, ErrTooLarge
	}
	return f.TMemoryBuffer.Write(buf)
}
