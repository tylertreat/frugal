package frugal

import (
	"encoding/binary"

	"git.apache.org/thrift.git/lib/go/thrift"
)

// TFramedMemoryBuffer implements TTransport using a bounded memory buffer.
// Writes which cause the buffer to exceed its size return ErrTooLarge.
// The TFramedMemoryBuffer handles framing data.
type TFramedMemoryBuffer struct {
	limit uint
	*thrift.TMemoryBuffer
}

var emptyFrameSize []byte = []byte{0, 0, 0, 0}

// NewTFramedMemoryBuffer returns a new TFramedMemoryBuffer with the given
// size limit. If the provided limit is non-positive, the buffer is allowed
// to grow unbounded.
func NewTFramedMemoryBuffer(size uint) *TFramedMemoryBuffer {
	buffer := &TFramedMemoryBuffer{size, thrift.NewTMemoryBuffer()}
	buffer.Write(emptyFrameSize)
	return buffer
}

// Write the data to the buffer. Returns ErrTooLarge if the write would cause
// the buffer to exceed its limit.
func (f *TFramedMemoryBuffer) Write(buf []byte) (int, error) {
	if f.limit > 0 && uint(len(buf)+f.Len()) > f.limit {
		f.Reset()
		return 0, ErrTooLarge
	}
	return f.TMemoryBuffer.Write(buf)
}

// Reset clears the buffer
func (f *TFramedMemoryBuffer) Reset() {
	f.TMemoryBuffer.Reset()
	f.Write(emptyFrameSize)
}

// Bytes retrieves the framed contents of the buffer.
func (f *TFramedMemoryBuffer) Bytes() []byte {
	data := f.TMemoryBuffer.Bytes()
	binary.BigEndian.PutUint32(data, uint32(len(data)-4))
	return data
}

// HasWriteData determines if there's any data in the buffer to send.
func (f *TFramedMemoryBuffer) HasWriteData() bool {
	return f.Len() > 4
}
