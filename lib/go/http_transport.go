package frugal

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"net/http"

	"git.apache.org/thrift.git/lib/go/thrift"
)

// NewFugalHandlerFunc is a function that create a ready to use Frugal handler function
func NewFrugalHandlerFunc(processor FProcessor, inPfactory, outPfactory *FProtocolFactory) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/x-thrift")
		buf := make([]byte, 4)
		decoder := base64.NewDecoder(base64.StdEncoding, r.Body)

		// Read and discard frame size
		n := 0
		for n < 4 {
			if p, err := decoder.Read(buf[n:]); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			} else {
				n += p
			}
		}

		// Read and process frame
		input := thrift.NewStreamTransportR(decoder)
		outBuf := new(bytes.Buffer)
		output := &thrift.TMemoryBuffer{Buffer: outBuf}
		if err := processor.Process(inPfactory.GetProtocol(input), outPfactory.GetProtocol(output)); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Encode and send response
		outBytes := outBuf.Bytes()
		binary.BigEndian.PutUint32(buf, uint32(len(outBytes)))
		encoded := new(bytes.Buffer)
		encoder := base64.NewEncoder(base64.StdEncoding, encoded)
		if _, err := encoder.Write(buf); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := encoder.Write(outBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := encoder.Close(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(encoded.Bytes())
	}
}
