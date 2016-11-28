package frugal

import (
	"errors"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

// Ensures IsErrTooLarge correctly classifies errors.
func TestIsErrTooLarge(t *testing.T) {
	assert.True(t, IsErrTooLarge(thrift.NewTTransportException(TTRANSPORT_REQUEST_TOO_LARGE, "error")))
	assert.True(t, IsErrTooLarge(thrift.NewTTransportException(TTRANSPORT_RESPONSE_TOO_LARGE, "error")))
	assert.False(t, IsErrTooLarge(nil))
	assert.False(t, IsErrTooLarge(errors.New("error")))
	assert.False(t, IsErrTooLarge(thrift.NewTTransportException(thrift.NOT_OPEN, "error")))
	assert.False(t, IsErrTooLarge(thrift.NewTApplicationException(0, "error")))
}
