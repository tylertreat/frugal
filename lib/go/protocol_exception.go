package frugal

import (
	"encoding/base64"

	"git.apache.org/thrift.git/lib/go/thrift"
)

// FProtocolException is a Frugal protocol error.
type FProtocolException interface {
	thrift.TProtocolException
}

type fProtocolException struct {
	typeId  int
	message string
}

// TypeId returns the exception type.
func (f *fProtocolException) TypeId() int {
	return f.typeId
}

// String returns a human-readable version of the exception.
func (f *fProtocolException) String() string {
	return f.message
}

// Error returns a human-readable error message.
func (f *fProtocolException) Error() string {
	return f.message
}

// NewFProtocolException wraps the given error in an FProtocolException.
func NewFProtocolException(err error) FProtocolException {
	if err == nil {
		return nil
	}
	if e, ok := err.(FProtocolException); ok {
		return e
	}
	if _, ok := err.(base64.CorruptInputError); ok {
		return &fProtocolException{thrift.INVALID_DATA, err.Error()}
	}
	return &fProtocolException{thrift.UNKNOWN_PROTOCOL_EXCEPTION, err.Error()}
}

// NewFProtocolExceptionWithType creates a new FProtocolException for the
// given type.
func NewFProtocolExceptionWithType(errType int, message string) FProtocolException {
	return &fProtocolException{errType, message}
}
