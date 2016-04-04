package frugal

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

var basicFrame = []byte{0, 0, 0, 0, 14, 0, 0, 0, 3, 102, 111, 111, 0, 0, 0, 3, 98, 97, 114}
var basicHeaders = map[string]string{"foo": "bar"}

var frugalFrame = []byte{0, 0, 0, 0, 65, 0, 0, 0, 5, 104, 101, 108, 108, 111, 0, 0, 0, 5,
	119, 111, 114, 108, 100, 0, 0, 0, 5, 95, 111, 112, 105, 100, 0, 0, 0, 1, 48, 0, 0, 0,
	4, 95, 99, 105, 100, 0, 0, 0, 21, 105, 89, 65, 71, 67, 74, 72, 66, 87, 67, 75, 76, 74,
	66, 115, 106, 107, 100, 111, 104, 98}
var frugalHeaders = map[string]string{opID: "0", cid: "iYAGCJHBWCKLJBsjkdohb", "hello": "world"}

var tProtocolFactory = thrift.NewTBinaryProtocolFactoryDefault()

func TestReadRequestHeaderMissingOpID(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(basicFrame)}
	proto := &FProtocol{tProtocolFactory.GetProtocol(transport)}

	expectedErr := NewFProtocolExceptionWithType(thrift.INVALID_DATA, "frugal: request missing op id")
	_, err := proto.ReadRequestHeader()
	assert.Equal(expectedErr, err)
}

func TestReadRequestHeader(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(frugalFrame)}
	proto := &FProtocol{tProtocolFactory.GetProtocol(transport)}

	ctx, err := proto.ReadRequestHeader()
	assert.Nil(err)
	assert.Equal(frugalHeaders[cid], ctx.CorrelationID())
	assert.Equal(uint64(0), ctx.opID())
	val, ok := ctx.RequestHeader("hello")
	assert.True(ok)
	assert.Equal(frugalHeaders["hello"], val)
}

func TestReadResponseHeader(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(basicFrame)}
	proto := &FProtocol{tProtocolFactory.GetProtocol(transport)}

	ctx := NewFContext("")
	proto.ReadResponseHeader(ctx)
	val, ok := ctx.ResponseHeader("foo")
	assert.True(ok)
	assert.Equal(val, "bar")
}

func TestWriteHeaderErroredWrite(t *testing.T) {
	assert := assert.New(t)
	mft := &mockFTransport{}
	writeErr := errors.New("write falied")
	mft.On("Write", basicFrame).Return(0, writeErr)
	proto := &FProtocol{tProtocolFactory.GetProtocol(mft)}
	expectedErr := thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION, fmt.Sprintf("frugal: error writing protocol headers: %s", writeErr))
	assert.Equal(expectedErr, proto.writeHeader(basicHeaders))
}

func TestWriteHeaderBadWrite(t *testing.T) {
	assert := assert.New(t)
	mft := &mockFTransport{}
	mft.On("Write", basicFrame).Return(0, nil)
	proto := &FProtocol{tProtocolFactory.GetProtocol(mft)}
	expectedErr := thrift.NewTTransportException(thrift.UNKNOWN_PROTOCOL_EXCEPTION, "frugal: failed to write complete protocol headers")
	assert.Equal(expectedErr, proto.writeHeader(basicHeaders))
}

func TestWriteHeader(t *testing.T) {
	assert := assert.New(t)
	mft := &mockFTransport{}
	mft.On("Write", basicFrame).Return(len(basicFrame), nil)
	proto := &FProtocol{tProtocolFactory.GetProtocol(mft)}
	assert.Nil(proto.writeHeader(basicHeaders))
}

func TestReadHeaderTransportError(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer([]byte{0})}
	_, err := readHeader(transport)
	assert.Error(err)
}

func TestReadHeaderUnsupportedVersion(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer([]byte{0x01, 0, 0, 0, 0})}
	expectedErr := NewFProtocolExceptionWithType(thrift.BAD_VERSION, "frugal: unsupported protocol version 1")
	_, err := readHeader(transport)
	assert.Equal(expectedErr, err)
}

func TestReadHeaderBadLength(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer([]byte{protocolV0, 0, 0, 0, 1})}
	_, err := readHeader(transport)
	assert.Error(err)
}

func TestReadHeader(t *testing.T) {
	assert := assert.New(t)
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(basicFrame)}

	headers, err := readHeader(transport)
	assert.Nil(err)
	assert.Equal(basicHeaders, headers)
}

func TestGetHeadersFromFrameInvalidSize(t *testing.T) {
	assert := assert.New(t)
	expectedErr := NewFProtocolExceptionWithType(thrift.INVALID_DATA, "frugal: invalid frame size 1")
	_, err := getHeadersFromFrame([]byte{0})
	assert.Equal(expectedErr, err)
}

func TestGetHeadersFromFrameUnsupportedVersion(t *testing.T) {
	assert := assert.New(t)
	expectedErr := NewFProtocolExceptionWithType(thrift.BAD_VERSION, "frugal: unsupported protocol version 1")
	_, err := getHeadersFromFrame([]byte{0x01, 0, 0, 0, 0})
	assert.Equal(expectedErr, err)
}

func TestGetHeadersFromFrame(t *testing.T) {
	assert := assert.New(t)
	headers, err := getHeadersFromFrame(basicFrame)
	assert.Nil(err)
	assert.Equal(basicHeaders, headers)
}
