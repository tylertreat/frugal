package frugal

import (
	"errors"
	"git.apache.org/thrift.git/lib/go/thrift"
)

// TTransportExceptions
const (
	// TTRANSPORT_REQUEST_TOO_LARGE is a TTransportException error type
	// indicating the request exceeded the size limit.
	TTRANSPORT_REQUEST_TOO_LARGE = 100

	// TTRANSPORT_RESPONSE_TOO_LARGE is a TTransportException error type
	// indicating the response exceeded the size limit.
	TTRANSPORT_RESPONSE_TOO_LARGE = 101
)

// TApplicationExceptions
const (
	// TAPPLICATION_RESPONSE_TOO_LARGE is a TApplicationException error type
	// indicating the response exceeded the size limit.
	TAPPLICATION_RESPONSE_TOO_LARGE = 100

	// TAPPLICATION_RATE_LIMIT_EXCEEDED is a TApplicationException error type
	// indicating the client exceeded its rate limit.
	TAPPLICATION_RATE_LIMIT_EXCEEDED = 102
)

var (
	// ErrTransportClosed is returned by service calls when the transport is
	// unexpectedly closed, perhaps as a result of the transport entering an
	// invalid state. If this is returned, the transport should be
	// reinitialized.
	ErrTransportClosed = errors.New("frugal: transport was unexpectedly closed")

	// ErrRateLimitExceeded is returned when the client exceeds its rate limit.
	ErrRateLimitExceeded = thrift.NewTApplicationException(
		TAPPLICATION_RATE_LIMIT_EXCEEDED,
		"client exceeded rate limit")
)

// IsErrTooLarge indicates if the given error is a TTransportException
// indicating an oversized request or response.
func IsErrTooLarge(err error) bool {
	if e, ok := err.(thrift.TTransportException); ok {
		return e.TypeId() == TTRANSPORT_REQUEST_TOO_LARGE || e.TypeId() == TTRANSPORT_RESPONSE_TOO_LARGE
	}
	return false
}
