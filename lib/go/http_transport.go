package frugal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"git.apache.org/thrift.git/lib/go/thrift"
)

var newEncoder = func(buf *bytes.Buffer) io.WriteCloser {
	return base64.NewEncoder(base64.StdEncoding, buf)
}

// NewFugalHandlerFunc is a function that create a ready to use Frugal handler function
func NewFrugalHandlerFunc(processor FProcessor, inPfactory, outPfactory *FProtocolFactory) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/x-frugal")

		// Read and process frame
		decoder := base64.NewDecoder(base64.StdEncoding, r.Body)
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

		// Encode and send response
		encoded := new(bytes.Buffer)
		encoder := newEncoder(encoded)
		if _, err := encoder.Write(outBuf.Bytes()); err != nil {
			http.Error(w,
				fmt.Sprintf("Problem encoding frugal bytes to base64 %s", err),
				http.StatusInternalServerError,
			)
			return
		}
		if err := encoder.Close(); err != nil {
			http.Error(w,
				fmt.Sprintf("Problem encoding frugal bytes to base64 %s", err),
				http.StatusInternalServerError,
			)
			return
		}
		w.Write(encoded.Bytes())
	}
}
