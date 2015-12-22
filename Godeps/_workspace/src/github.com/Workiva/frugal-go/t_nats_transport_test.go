package frugal

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/nats-io/gnatsd/server"
	"github.com/nats-io/nats"
	"github.com/stretchr/testify/assert"
)

// Ensure Open subscribes to the set subject and writes received bytes to the
// writer which are returned on Read. Write writes bytes to the buffer which
// is sent over NATS on Flush.
func TestNATSTransport(t *testing.T) {
	assert := assert.New(t)
	s := runServer(nil)
	defer s.Shutdown()

	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", port))
	if err != nil {
		t.Fatalf("Error creating client: %v\n", err)
	}
	transport := newTNatsTransport(conn)
	transport.SetSubject("foo")
	assert.False(transport.IsOpen())

	assert.Nil(transport.Open())

	data := []byte("this is a test")
	n, err := transport.Write(data)
	assert.Nil(err)
	assert.Equal(len(data), n)

	assert.Nil(transport.Flush())

	buff := make([]byte, len(data))
	n, err = transport.Read(buff)
	assert.Nil(err)
	assert.True(bytes.Equal(buff, data))

	assert.True(transport.IsOpen())

	transport.Close()
	assert.False(transport.IsOpen())

	data = []byte("another test")
	n, err = transport.Write(data)
	assert.Nil(err)
	assert.Equal(len(data), n)

	assert.Nil(transport.Flush())

	buff = make([]byte, len(data))
	n, err = transport.Read(buff)
	assert.Error(err)
	assert.False(bytes.Equal(buff, data))

	assert.Equal(^uint64(0), transport.RemainingBytes())

	bigMessage := make([]byte, 1024*1024+1)
	n, err = transport.Write(bigMessage)
	assert.Equal(0, n)
	assert.Equal(ErrTooLarge, err)

	assert.Equal(ErrEmptyMessage, transport.Flush())
}

const port = 11222

var defaultOptions = server.Options{
	Host:        "localhost",
	Port:        port,
	HTTPPort:    11333,
	ClusterPort: 11444,
	ProfPort:    11280,
	NoLog:       true,
	NoSigs:      true,
}

// New Go Routine based server
func runServer(opts *server.Options) *server.Server {
	if opts == nil {
		opts = &defaultOptions
	}
	s := server.New(opts)
	if s == nil {
		panic("No NATS Server object returned.")
	}

	// Run server in Go routine.
	go s.Start()

	end := time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		addr := s.Addr()
		if addr == nil {
			time.Sleep(10 * time.Millisecond)
			// Retry. We might take a little while to open a connection.
			continue
		}
		conn, err := net.Dial("tcp", addr.String())
		if err != nil {
			// Retry after 50ms
			time.Sleep(50 * time.Millisecond)
			continue
		}
		conn.Close()
		return s
	}
	panic("Unable to start NATS Server in Go Routine")

}
