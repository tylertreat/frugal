package frugal

import (
	"fmt"
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Ensures Open returns an error if NATS is not connected.
func TestStatelessNatsTransportOpenNatsDisconnected(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	tr := NewStatelessNatsTTransport(conn, "foo", "bar")

	assert.Error(t, tr.Open())
	assert.False(t, tr.IsOpen())
}

// Ensures Open returns an ALREADY_OPEN TTransportException if the transport
// is already open.
func TestStatelessNatsTransportOpenAlreadyOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newStatelessClientAndServer(t)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	err := tr.Open()
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.ALREADY_OPEN, trErr.TypeId())
}

// Ensures Open subscribes to the right subject and buffers received frames
// which are returned on calls to Read.
func TestStatelessNatsTransportOpenRead(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newStatelessClientAndServer(t)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	frame := []byte("helloworld")
	assert.Nil(t, conn.Publish(tr.inbox, frame))

	buff := make([]byte, 5)
	n, err := tr.Read(buff)
	assert.Nil(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte("hello"), buff)
	n, err = tr.Read(buff)
	assert.Nil(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte("world"), buff)
}

// Ensures Read returns a NOT_OPEN TTransportException if the transport is not
// open.
func TestStatelessNatsTransportReadNotOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	tr := NewStatelessNatsTTransport(conn, "foo", "bar")

	n, err := tr.Read(make([]byte, 5))
	assert.Equal(t, 0, n)
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Read returns a NOT_OPEN TTransportException if NATS is not connected.
func TestStatelessNatsTransportReadNatsDisconnected(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newStatelessClientAndServer(t)
	defer server.Stop()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	conn.Close()

	buff := make([]byte, 5)
	n, err := tr.Read(buff)
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
	assert.Equal(t, 0, n)
}

// Ensures Write returns a NOT_OPEN TTransportException if the transport is not
// open.
func TestStatelessNatsTransportWriteNotOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	tr := NewStatelessNatsTTransport(conn, "foo", "bar")

	n, err := tr.Write(make([]byte, 10))

	assert.Equal(t, 0, n)
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Write returns a NOT_OPEN TTransportException if NATS is not
// connected.
func TestStatelessNatsTransportWriteNatsDisconnected(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newStatelessClientAndServer(t)
	defer server.Stop()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	conn.Close()

	buff := make([]byte, 5)
	n, err := tr.Write(buff)
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
	assert.Equal(t, 0, n)
}

// Ensures Write buffers data. If the buffer exceeds 1MB, ErrTooLarge is
// returned.
func TestStatelessNatsTransportWrite(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newStatelessClientAndServer(t)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	buff := make([]byte, 5)
	n, err := tr.Write(buff)
	assert.Nil(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 5, tr.requestBuffer.Len())
	buff = make([]byte, 1024*1024)
	n, err = tr.Write(buff)
	assert.Equal(t, ErrTooLarge, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, 0, tr.requestBuffer.Len())
}

// Ensures Flush returns a NOT_OPEN TTransportException if the transport is not
// open.
func TestStatelessNatsTransportFlushNotOpen(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	tr := NewStatelessNatsTTransport(conn, "foo", "bar")

	err = tr.Flush()
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Flush returns a NOT_OPEN TTransportException if NATS is not
// connected.
func TestStatelessNatsTransportFlushNatsDisconnected(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newStatelessClientAndServer(t)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	conn.Close()

	err := tr.Flush()
	trErr := err.(thrift.TTransportException)
	assert.Equal(t, thrift.NOT_OPEN, trErr.TypeId())
}

// Ensures Flush doesn't send anything to NATS if no data is buffered.
func TestStatelessNatsTransportFlushNoData(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newStatelessClientAndServer(t)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	sub, err := conn.SubscribeSync(tr.subject)
	assert.Nil(t, err)
	assert.Nil(t, tr.Flush())
	conn.Flush()
	_, err = sub.NextMsg(5 * time.Millisecond)
	assert.Equal(t, nats.ErrTimeout, err)
}

// Ensures Flush sends the frame to the correct NATS subject.
func TestStatelessNatsTransportFlush(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	tr, server, conn := newStatelessClientAndServer(t)
	defer server.Stop()
	defer conn.Close()
	assert.Nil(t, tr.Open())
	defer tr.Close()
	assert.True(t, tr.IsOpen())

	frame := []byte("helloworld")
	_, err := tr.Write(frame)
	assert.Nil(t, err)
	sub, err := conn.SubscribeSync(tr.subject)
	assert.Nil(t, err)
	assert.Nil(t, tr.Flush())
	conn.Flush()
	msg, err := sub.NextMsg(5 * time.Millisecond)
	assert.Nil(t, err)
	assert.Equal(t, frame, msg.Data)
}

func newStatelessClientAndServer(t *testing.T) (*statelessNatsTTransport, *FStatelessNatsServer, *nats.Conn) {
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	mockProcessor := new(mockFProcessor)
	mockTProtocolFactory := new(mockTProtocolFactory)
	protocolFactory := NewFProtocolFactory(mockTProtocolFactory)
	server := NewFStatelessNatsServer(conn, mockProcessor, protocolFactory,
		protocolFactory, "foo", "queue", 1)
	mockTransport := new(mockFTransport)
	proto := thrift.NewTJSONProtocol(mockTransport)
	mockTProtocolFactory.On("GetProtocol", mock.AnythingOfType("*thrift.TMemoryBuffer")).Return(proto)
	fproto := &FProtocol{proto}
	mockProcessor.On("Process", fproto, fproto).Return(nil)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)
	tr := NewStatelessNatsTTransport(conn, "foo", "bar")
	return tr.(*statelessNatsTTransport), server.(*FStatelessNatsServer), conn
}
