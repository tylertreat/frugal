package frugal

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

type mockFProcessorForHttp struct {
	err             error
	expectedPayload []byte
	response        []byte
}

func (m *mockFProcessorForHttp) Process(iprot, oprot *FProtocol) error {
	if m.expectedPayload != nil {
		actual := make([]byte, len(m.expectedPayload))
		if _, err := io.ReadFull(iprot.TProtocol.Transport(), actual); err != nil {
			return err
		}
		for i := range m.expectedPayload {
			if actual[i] != m.expectedPayload[i] {
				return errors.New("Payload doesn't match expected")
			}
		}
	}

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

// Ensures that the payload size header has the wrong format an error is
// returned
func TestFrugalHandlerFuncHeaderError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	r, err := http.NewRequest("POST", "fooUrl", nil)
	r.Header.Add(payloadLimitHeader, "foo")
	assert.Nil(err)

	mockProcessor := &mockFProcessorForHttp{}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusBadRequest)
	assert.Equal(
		[]byte(fmt.Sprintf("%s header not an integer\n", payloadLimitHeader)),
		w.Body.Bytes(),
	)
}

// Ensures that if there is an error reading the frame size out of the request
// payload, an error is returned
func TestFrugalHandlerFuncFrameSizeError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	framedBody := []byte{0, 1, 2}
	encodedBody := base64.StdEncoding.EncodeToString(framedBody)
	r, err := http.NewRequest("POST", "fooUrl", strings.NewReader(encodedBody))
	assert.Nil(err)

	mockProcessor := &mockFProcessorForHttp{}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusBadRequest)
	assert.Equal(
		[]byte(fmt.Sprintf("Could not read the frugal frame bytes %s\n", io.ErrUnexpectedEOF)),
		w.Body.Bytes(),
	)
}

// Ensures that processor errors are handled and routed back in the http
// response
func TestFrugalHandlerFuncProcessorError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	expectedBody := []byte{4, 5, 6, 7, 8}
	framedBody := append([]byte{0, 1, 2, 3}, expectedBody...)
	encodedBody := base64.StdEncoding.EncodeToString(framedBody)
	r, err := http.NewRequest("POST", "fooUrl", strings.NewReader(encodedBody))
	assert.Nil(err)

	processorErr := fmt.Errorf("processor error")
	mockProcessor := &mockFProcessorForHttp{expectedPayload: expectedBody, err: processorErr}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusBadRequest)
	assert.Equal(
		[]byte(fmt.Sprintf("Frugal request failed %s\n", processorErr)),
		w.Body.Bytes(),
	)
}

// Ensures that if the response payload exceeds the request limit, a
// RequestEntityTooLarge error is returned.
func TestFrugalHandlerFuncTooLargeError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	expectedBody := []byte{4, 5, 6, 7, 8}
	framedBody := append([]byte{0, 1, 2, 3}, expectedBody...)
	encodedBody := base64.StdEncoding.EncodeToString(framedBody)
	r, err := http.NewRequest("POST", "fooUrl", strings.NewReader(encodedBody))
	r.Header.Add(payloadLimitHeader, "5")
	assert.Nil(err)

	response := make([]byte, 10)
	mockProcessor := &mockFProcessorForHttp{expectedPayload: expectedBody, response: response}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusRequestEntityTooLarge)
	assert.Equal(
		[]byte("Response size (10) larger than requested size (5)\n"),
		w.Body.Bytes(),
	)
}

// Ensures that base64 encoding errors are handled and routed back in the http
// response
func TestFrugalHandlerFuncBase64WriteError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()

	expectedBody := []byte{4, 5, 6, 7, 8}
	framedBody := append([]byte{0, 1, 2, 3}, expectedBody...)
	encodedBody := base64.StdEncoding.EncodeToString(framedBody)
	r, err := http.NewRequest("POST", "fooUrl", strings.NewReader(encodedBody))
	assert.Nil(err)

	response := []byte("Hello")
	mockProcessor := &mockFProcessorForHttp{expectedPayload: expectedBody, response: response}
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

	expectedBody := []byte{4, 5, 6, 7, 8}
	framedBody := append([]byte{0, 1, 2, 3}, expectedBody...)
	encodedBody := base64.StdEncoding.EncodeToString(framedBody)
	r, err := http.NewRequest("POST", "fooUrl", strings.NewReader(encodedBody))
	assert.Nil(err)

	response := []byte("Hello")
	mockProcessor := &mockFProcessorForHttp{expectedPayload: expectedBody, response: response}
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

	expectedBody := []byte{4, 5, 6, 7, 8}
	framedBody := append([]byte{0, 1, 2, 3}, expectedBody...)
	encodedBody := base64.StdEncoding.EncodeToString(framedBody)
	r, err := http.NewRequest("POST", "fooUrl", strings.NewReader(encodedBody))
	assert.Nil(err)

	response := []byte{9, 10, 11, 12}
	mockProcessor := &mockFProcessorForHttp{expectedPayload: expectedBody, response: response}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	handler := NewFrugalHandlerFunc(mockProcessor, protocolFactory, protocolFactory)

	handler(w, r)

	assert.Equal(w.Code, http.StatusOK)
	assert.Equal(w.Header().Get("content-type"), "application/x-frugal")
	assert.Equal(w.Header().Get("content-transfer-encoding"), "base64")
	assert.Equal(
		[]byte(base64.StdEncoding.EncodeToString(append([]byte{0, 0, 0, 4}, response...))),
		w.Body.Bytes(),
	)

}
