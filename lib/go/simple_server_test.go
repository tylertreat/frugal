package frugal

import (
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/stretchr/mock"
	"github.com/stretchr/testify/assert"
)

const simpleServerAddr = "localhost:5535"

type mockFProcessorFactory struct {
	mock.Mock
}

func (m *mockFProcessorFactory) GetProcessor(tr thrift.TTransport) FProcessor {
	return m.Called(tr).Get(0).(FProcessor)
}

// Ensures FSimpleServer accepts connections.
func TestSimpleServer(t *testing.T) {
	mockFProcessorFactory := new(mockFProcessorFactory)
	protoFactory := thrift.NewTJSONProtocolFactory()
	fTransportFactory := NewFMuxTransportFactory(1)
	serverTr, err := thrift.NewTServerSocket(simpleServerAddr)
	if err != nil {
		t.Fatal(err)
	}
	server := NewFSimpleServerFactory5(
		mockFProcessorFactory,
		serverTr,
		NewFMuxTransportFactory(2),
		NewFProtocolFactory(protoFactory),
	)

	go func() {
		assert.Nil(t, server.Serve())
	}()
	time.Sleep(10 * time.Millisecond)

	mockFProcessor := new(mockFProcessor)
	mockFProcessorFactory.On("GetProcessor", mock.AnythingOfType("*thrift.TSocket")).Return(mockFProcessor)
	mockFProcessor.On("Process", mock.AnythingOfType("*frugal.FProtocol"),
		mock.AnythingOfType("*frugal.FProtocol")).Return(nil)

	transport, err := thrift.NewTSocket(simpleServerAddr)
	if err != nil {
		t.Fatal(err)
	}
	fTransport := fTransportFactory.GetTransport(transport)
	defer fTransport.Close()
	if err := fTransport.Open(); err != nil {
		t.Fatal(err)
	}

	_, err = fTransport.Write(make([]byte, 10))
	assert.Nil(t, err)
	assert.Nil(t, fTransport.Flush())
	time.Sleep(5 * time.Millisecond)

	assert.Nil(t, server.Stop())

	mockFProcessorFactory.AssertExpectations(t)
	mockFProcessor.AssertExpectations(t)
}
