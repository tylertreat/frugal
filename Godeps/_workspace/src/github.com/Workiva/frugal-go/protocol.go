package frugal

import "git.apache.org/thrift.git/lib/go/thrift"

// FProtocolFactory is a factory for FProtocol
type FProtocolFactory interface {
	GetProtocol(trans thrift.TTransport) FProtocol
}

// FProtocol is an extension of thrift TProtocol with the addition of headers
type FProtocol interface {
	thrift.TProtocol

	// WriteRequestHeader writes the request headers set on the given Context
	// into the protocol
	WriteRequestHeader(Context) error
	// ReadRequestHeader reads the request headers on the protocol into a
	// returned Context
	ReadRequestHeader() (Context, error)
	// WriteResponseHeader writes the response headers set on the given Context
	// into the protocol
	WriteResponseHeader(Context) error
	// ReadResponseHeader reads the response headers on the protocol into a
	// returned Context
	ReadResponseHeader(Context) error
}

// FBinaryProtocol is an implementation of FProtocol backed by the thrift
// binary protocol
type FBinaryProtocol struct {
	*thrift.TBinaryProtocol
	tr thrift.TTransport
}

// FBinaryProtocolFactory is a protocol factory backed by FBinary protocol
type FBinaryProtocolFactory struct {
	strictRead  bool
	strictWrite bool
}

// GetProtocol returns the FBinary protocol for the given thrift transport
func (p *FBinaryProtocolFactory) GetProtocol(t thrift.TTransport) FProtocol {
	return NewFBinaryProtocol(t, p.strictRead, p.strictWrite)
}

// NewFBinaryProtocolFactoryDefault returns the default implementation of
// FBinaryProtocolFactory - strict writes but not reads
func NewFBinaryProtocolFactoryDefault() *FBinaryProtocolFactory {
	return NewFBinaryProtocolFactory(false, true)
}

// NewFBinaryProtocolFactory returns an instance of FBinaryProtocolFactory
// with the given read/write rules
func NewFBinaryProtocolFactory(strictRead, strictWrite bool) *FBinaryProtocolFactory {
	return &FBinaryProtocolFactory{strictRead: strictRead, strictWrite: strictWrite}
}

// NewFBinaryProtocol returns an instance of FBinaryProtocol with the given
// thrift transport and read/write rules
func NewFBinaryProtocol(tr thrift.TTransport, strictRead, strictWrite bool) *FBinaryProtocol {
	return &FBinaryProtocol{
		thrift.NewTBinaryProtocol(tr, strictRead, strictWrite),
		tr,
	}
}

// WriteRequestHeader writes the request headers set on the given Context
// into the protocol
func (b *FBinaryProtocol) WriteRequestHeader(ctx Context) error {
	return b.writeHeader(ctx.RequestHeaders())
}

// ReadRequestHeader reads the request headers on the protocol into a
// returned Context
func (b *FBinaryProtocol) ReadRequestHeader() (Context, error) {
	// Check version when more are available
	_, err := b.ReadByte()
	if err != nil {
		return nil, err
	}
	numHeaders, err := b.ReadI16()
	if err != nil {
		return nil, err
	}

	ctx := &context{
		requestHeaders:  make(map[string]string),
		responseHeaders: make(map[string]string),
	}
	for i := int16(0); i < numHeaders; i++ {
		key, err := b.ReadString()
		if err != nil {
			return nil, err
		}
		value, err := b.ReadString()
		if err != nil {
			return nil, err
		}
		ctx.AddRequestHeader(key, value)
	}
	return ctx, nil
}

// WriteResponseHeader writes the response headers set on the given Context
// into the protocol
func (b *FBinaryProtocol) WriteResponseHeader(ctx Context) error {
	return b.writeHeader(ctx.ResponseHeaders())
}

// ReadResponseHeader reads the response headers on the protocol into a
// returned Context
func (b *FBinaryProtocol) ReadResponseHeader(ctx Context) error {
	// Check version when more are available
	_, err := b.ReadByte()
	if err != nil {
		return err
	}
	numHeaders, err := b.ReadI16()
	if err != nil {
		return err
	}
	for i := int16(0); i < numHeaders; i++ {
		key, err := b.ReadString()
		if err != nil {
			return err
		}
		value, err := b.ReadString()
		if err != nil {
			return err
		}
		ctx.AddResponseHeader(key, value)
	}
	return nil
}

func (b *FBinaryProtocol) writeHeader(headers map[string]string) error {
	// Write version
	if err := b.WriteByte(0x00); err != nil {
		return err
	}
	if err := b.WriteI16(int16(len(headers))); err != nil {
		return err
	}
	for key, value := range headers {
		if err := b.WriteString(key); err != nil {
			return err
		}
		if err := b.WriteString(value); err != nil {
			return err
		}
	}
	return nil
}
