package frugal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

type mockFProcessorForHttp struct {
	err      error
	response []byte
}

func (m *mockFProcessorForHttp) Process(iprot, oprot *FProtocol) error {
	if m.err != nil {
		return m.err
	}
	oprot.TProtocol.Transport().Write(m.response)
	return nil
}

type mockWriteCloser struct {
	writeErr error
	closeErr error
}

func (m *mockWriteCloser) Write(p []byte) (int, error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return len(p), nil
}

func (m *mockWriteCloser) Close() error {
	return m.closeErr
}

// Ensures that processor errors are handled and routed back in the http
// response
func TestFrugalHandlerFuncProcessorError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	r, err := http.NewRequest("POST", "fooUrl", nil)
	assert.Nil(err)

	processorErr := fmt.Errorf("processor error")
	mockProcessor := &mockFProcessorForHttp{err: processorErr}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusBadRequest)
	assert.Equal(
		[]byte(fmt.Sprintf("Frugal request failed %s\n", processorErr)),
		w.Body.Bytes(),
	)
}

// Ensures that base64 encoding errors are handled and routed back in the http
// response
func TestFrugalHandlerFuncBase64WriteError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	r, err := http.NewRequest("POST", "fooUrl", nil)
	assert.Nil(err)

	response := []byte("Hello")
	mockProcessor := &mockFProcessorForHttp{response: response}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	oldNewEncoder := newEncoder
	writeErr := fmt.Errorf("write err")
	newEncoder = func(_ *bytes.Buffer) io.WriteCloser {
		return &mockWriteCloser{writeErr: writeErr}
	}

	handler(w, r)

	assert.Equal(w.Code, http.StatusInternalServerError)
	assert.Equal(
		[]byte(fmt.Sprintf("Problem encoding frugal bytes to base64 %s\n", writeErr)),
		w.Body.Bytes(),
	)

	newEncoder = oldNewEncoder
}

// Ensures that base64 encoding errors are handled and routed back in the http
// response
func TestFrugalHandlerFuncBase64CloseError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	r, err := http.NewRequest("POST", "fooUrl", nil)
	assert.Nil(err)

	response := []byte("Hello")
	mockProcessor := &mockFProcessorForHttp{response: response}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	oldNewEncoder := newEncoder
	closeErr := fmt.Errorf("close err")
	newEncoder = func(_ *bytes.Buffer) io.WriteCloser {
		return &mockWriteCloser{closeErr: closeErr}
	}

	handler(w, r)

	assert.Equal(w.Code, http.StatusInternalServerError)
	assert.Equal(
		[]byte(fmt.Sprintf("Problem encoding frugal bytes to base64 %s\n", closeErr)),
		w.Body.Bytes(),
	)

	newEncoder = oldNewEncoder
}

// Ensures that the frugal payload is appropriately processed and returned
func TestFrugalHandlerFunc(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	r, err := http.NewRequest("POST", "fooUrl", nil)
	assert.Nil(err)

	response := []byte("Hello")
	mockProcessor := &mockFProcessorForHttp{response: response}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusOK)
	assert.Equal(
		[]byte(base64.StdEncoding.EncodeToString(response)),
		w.Body.Bytes(),
	)

}
