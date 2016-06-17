package frugal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensures Write writes to the buffer until its size limit is reached, after
// which ErrTooLarge is returned and the buffer is reset.
func TestFBoundedMemoryBufferWrite(t *testing.T) {
	buff := NewFBoundedMemoryBuffer(100)
	assert.Equal(t, 0, buff.Len())
	n, err := buff.Write(make([]byte, 50))
	assert.Nil(t, err)
	assert.Equal(t, 50, n)
	n, err = buff.Write(make([]byte, 40))
	assert.Nil(t, err)
	assert.Equal(t, 40, n)
	assert.Equal(t, 90, buff.Len())
	_, err = buff.Write(make([]byte, 20))
	assert.Equal(t, ErrTooLarge, err)
	assert.Equal(t, 0, buff.Len())
}
