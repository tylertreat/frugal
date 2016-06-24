package frugal

import (
	"fmt"
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
	"github.com/stretchr/testify/assert"
)

// Ensures FStatelessNatsServer receives requests and sends back responses on
// the correct subject.
func TestFStatelessNatsServer(t *testing.T) {
	s := runServer(nil)
	defer s.Shutdown()
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", defaultOptions.Port))
	if err != nil {
		t.Fatal(err)
	}
	processor := &processor{t}
	protoFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	server := NewFStatelessNatsServerBuilder(conn, processor, protoFactory, "foo").WithQueueGroup("queue").Build()
	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)
	defer server.Stop()

	tr := NewStatelessNatsTTransport(conn, "foo", "bar")
	assert.Nil(t, tr.Open())

	// Send a request.
	_, err = tr.Write([]byte("xxxxhello"))
	assert.Nil(t, err)
	assert.Nil(t, tr.Flush())

	// Get the response.
	buff := make([]byte, 100)
	n, err := tr.Read(buff)
	assert.Nil(t, err)

	// Server should send back "foo" in binary protocol.
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0, 0x3, 0x66, 0x6f, 0x6f}, buff[0:n])
}

type processor struct {
	t *testing.T
}

func (p *processor) Process(in, out *FProtocol) error {
	assert.Equal(p.t, "hello", string(in.Transport().(*thrift.TMemoryBuffer).Bytes()))
	out.WriteString("foo")
	return nil
}
