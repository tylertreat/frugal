package frugal

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
)

const payloadLimitHeader = "x-frugal-payload-limit"

var newEncoder = func(buf *bytes.Buffer) io.WriteCloser {
	return base64.NewEncoder(base64.StdEncoding, buf)
}

// NewFugalHandlerFunc is a function that create a ready to use Frugal handler function
func NewFrugalHandlerFunc(processor FProcessor, inPfactory, outPfactory *FProtocolFactory) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/x-frugal")

		// Check for size limitation
		limitStr := r.Header.Get(payloadLimitHeader)
		var limit int64 = 0
		if limitStr != "" {
			var err error
			limit, err = strconv.ParseInt(limitStr, 10, 64)
			if err != nil {
				http.Error(w,
					fmt.Sprintf("%s header not an integer", payloadLimitHeader),
					http.StatusBadRequest,
				)
				return
			}
		}

		// Create a decoder based on the payload
		decoder := base64.NewDecoder(base64.StdEncoding, r.Body)

		// Read out the frame size
		frameSize := make([]byte, 4)
		if _, err := io.ReadFull(decoder, frameSize); err != nil {
			http.Error(w,
				fmt.Sprintf("Could not read the frugal frame bytes %s", err),
				http.StatusBadRequest,
			)
			return
		}

		// Read and process frame
		input := thrift.NewStreamTransportR(decoder)
		outBuf := new(bytes.Buffer)
		output := &thrift.TMemoryBuffer{Buffer: outBuf}
		if err := processor.Process(inPfactory.GetProtocol(input), outPfactory.GetProtocol(output)); err != nil {
			http.Error(w,
				fmt.Sprintf("Frugal request failed %s", err),
				http.StatusBadRequest,
			)
			return
		}

		// If client requested a limit, check the buffer size
		if limit > 0 && outBuf.Len() > int(limit) {
			http.Error(w,
				fmt.Sprintf("Response size (%d) larger than requested size (%d)", outBuf.Len(), limit),
				http.StatusRequestEntityTooLarge,
			)
			return
		}

		// Encode response
		encoded := new(bytes.Buffer)
		encoder := newEncoder(encoded)
		var err error
		binary.BigEndian.PutUint32(frameSize, uint32(outBuf.Len()))
		if _, e := encoder.Write(frameSize); e != nil {
			err = e
		}
		if _, e := encoder.Write(outBuf.Bytes()); e != nil {
			err = e
		}
		if e := encoder.Close(); e != nil {
			err = e
		}

		// Check for encoding errors
		if err != nil {
			http.Error(w,
				fmt.Sprintf("Problem encoding frugal bytes to base64 %s", err),
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Add("content-transfer-encoding", "base64")
		w.Write(encoded.Bytes())
	}
}

// httpTTransport implements thrift.TTransport. This is a "stateless"
// transport in the sense that there is no connection with a server. A request
// is simply an http request and a response is an http response. This assumes
// requests/responses fit within a single http request.
type httpTTransport struct {
	client            *http.Client
	url               string
	requestSizeLimit  uint
	responseSizeLimit uint
	frameBuffer       chan []byte
	currentFrame      []byte
	requestBuffer     *bytes.Buffer
	mu                sync.RWMutex
	closeChan         chan struct{}
	isOpen            bool
}

// NewHttpTTransport returns a new Thrift TTransport which uses the
// HTTP as the underlying transport. This TTransport is stateless in that
// there is no connection maintained between the client and server. A request
// is simply an http request and a response is an http response. This assumes
// requests/responses fit within a single http request.
func NewHttpTTransport(client *http.Client, url string) thrift.TTransport {
	return &httpTTransport{
		client:        client,
		url:           url,
		frameBuffer:   make(chan []byte, frameBufferSize),
		requestBuffer: new(bytes.Buffer),
	}
}

// Open initializes the close channel and sets the open flag to true.
func (h *httpTTransport) Open() error {
	h.mu.Lock()
	h.closeChan = make(chan struct{})
	h.isOpen = true
	h.mu.Unlock()
	return nil
}

func (h *httpTTransport) handler(data []byte) {
	select {
	case h.frameBuffer <- data:
	case <-h.closeChan:
	}
}

func (h *httpTTransport) IsOpen() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isOpen
}

// Close closes the close channel and sets the open flag to false.
func (h *httpTTransport) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.isOpen {
		return nil
	}
	close(h.closeChan)
	h.isOpen = false
	return nil
}

func (h *httpTTransport) Read(buf []byte) (int, error) {
	if !h.IsOpen() {
		return 0, h.getClosedConditionError("read:")
	}
	if h.currentFrame == nil {
		select {
		case frame := <-h.frameBuffer:
			h.currentFrame = frame
		case <-h.closeChan:
			return 0, thrift.NewTTransportExceptionFromError(io.EOF)
		}
	}
	num := copy(buf, h.currentFrame)
	// TODO: We could be more efficient here. If the provided buffer isn't
	// full, we could attempt to get the next frame.

	h.currentFrame = h.currentFrame[num:]
	if len(h.currentFrame) == 0 {
		// The entire frame was copied, clear it.
		h.currentFrame = nil
	}
	return num, nil
}

// Write the bytes to a buffer. Returns ErrTooLarge if the buffer exceeds the
// client specified request size limit.
func (h *httpTTransport) Write(buf []byte) (int, error) {
	if !h.IsOpen() {
		return 0, h.getClosedConditionError("write:")
	}
	if h.requestSizeLimit > 0 && uint(len(buf)+h.requestBuffer.Len()) > h.requestSizeLimit {
		h.requestBuffer.Reset()
		return 0, ErrTooLarge
	}
	num, err := h.requestBuffer.Write(buf)
	return num, thrift.NewTTransportExceptionFromError(err)
}

// Flush sends the buffered bytes over HTTP.
func (h *httpTTransport) Flush() error {
	if !h.IsOpen() {
		return h.getClosedConditionError("flush:")
	}
	defer h.requestBuffer.Reset()
	data := h.requestBuffer.Bytes()
	if len(data) == 0 {
		return nil
	}

	response, err := h.makeRequest(data)
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}

	// All responses should be framed with 4 bytes (uint32)
	if len(response) < 4 {
		return thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
			errors.New("frugal: invalid frame size"))
	}

	// If there are only 4 bytes, this needs to be a one-way
	// (i.e. frame size 0)
	if len(response) == 4 {
		for i := range response {
			if response[i] != 0 {
				return thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA,
					errors.New("frugal: missing data"))
			}
		}
		// it's a one-way, drop it
		return nil
	}

	select {
	case h.frameBuffer <- response:
	case <-h.closeChan:
	}

	return nil
}

func (h *httpTTransport) makeRequest(requestPayload []byte) ([]byte, error) {
	// Encode request payload
	encoded := new(bytes.Buffer)
	encoder := newEncoder(encoded)
	if _, err := encoder.Write(requestPayload); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}

	// Initialize request
	request, err := http.NewRequest("POST", h.url, encoded)
	if err != nil {
		return nil, err
	}

	// Add request headers
	request.Header.Add("content-type", "application/x-frugal")
	request.Header.Add("content-transfer-encoding", "base64")
	request.Header.Add("accept", "application/x-frugal")
	if h.responseSizeLimit > 0 {
		request.Header.Add(payloadLimitHeader, strconv.FormatUint(uint64(h.responseSizeLimit), 10))
	}

	// Make request
	response, err := h.client.Do(request)
	if err != nil {
		return nil, err
	}

	// Response too large
	if response.StatusCode == http.StatusRequestEntityTooLarge {
		return nil, thrift.NewTTransportException(RESPONSE_TOO_LARGE,
			"request was too large for the transport")
	}

	// Decode body
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(response.Body); err != nil {
		return nil, err
	}
	body := string(buf.Bytes())

	// Check bad status code
	if response.StatusCode >= 300 {
		return nil, errors.New(body)
	}

	// Decode and return response body
	bts, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return nil, err
	}
	fmt.Println(bts)
	return bts, nil

}

func (h *httpTTransport) RemainingBytes() uint64 {
	return ^uint64(0)
}

func (h *httpTTransport) getClosedConditionError(prefix string) error {
	return thrift.NewTTransportException(thrift.NOT_OPEN,
		fmt.Sprintf("%s HTTP TTransport not open", prefix))
}
