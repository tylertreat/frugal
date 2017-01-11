package frugal

import (
	"io"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
	"time"
)

type fAdapterTransportFactory struct{}

// NewAdapterTransportFactory creates a new FTransportFactory which produces an
// FTransport implementation that acts as an adapter for thrift.TTransport.
// This allows TTransports which support blocking reads to work with Frugal by
// starting a goroutine that reads from the underlying transport and calling
// the registry on received frames.
func NewAdapterTransportFactory() FTransportFactory {
	return &fAdapterTransportFactory{}
}

// GetTransport returns a new adapter FTransport.
func (f *fAdapterTransportFactory) GetTransport(tr thrift.TTransport) FTransport {
	return NewAdapterTransport(tr)
}

type fAdapterTransport struct {
	transport          thrift.TTransport
	isOpen             bool
	mu                 sync.RWMutex
	closeSignal        chan struct{}
	closeChan          chan error
	monitorCloseSignal chan<- error
	registry           FRegistry
}

// NewAdapterTransport returns an FTransport which uses the given TTransport
// for read/write operations in a way that is compatible with Frugal. This
// allows TTransports which support blocking reads to work with Frugal by
// starting a goroutine that reads from the underlying transport and calling
// the registry on received frames.
func NewAdapterTransport(tr thrift.TTransport) FTransport {
	return &fAdapterTransport{
		registry:    NewFRegistry(),
		transport:   tr,
		closeSignal: make(chan struct{}, 1),
	}
}

// Open prepares the transport to send data.
func (f *fAdapterTransport) Open() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.isOpen {
		return thrift.NewTTransportException(thrift.ALREADY_OPEN,
			"frugal: transport already open")
	}

	if err := f.transport.Open(); err != nil {
		// It's OK if the underlying transport is already open.
		if e, ok := err.(thrift.TTransportException); !(ok && e.TypeId() == thrift.ALREADY_OPEN) {
			return err
		}
	}

	go f.readLoop()
	f.isOpen = true
	f.closeChan = make(chan error, 1)
	return nil
}

func (f *fAdapterTransport) readLoop() {
	framedTransport := NewTFramedTransport(f.transport)
	for {
		frame, err := f.readFrame(framedTransport)
		if err != nil {
			// First check if the transport was closed.
			select {
			case <-f.closeSignal:
				// Transport was closed.
				return
			default:
			}

			if err, ok := err.(thrift.TTransportException); ok && err.TypeId() == thrift.END_OF_FILE {
				// EOF indicates remote peer disconnected.
				f.Close()
				return
			}

			logger().Error("frugal: error reading protocol frame, closing transport: ", err)
			f.close(err)
			return
		}

		if err := f.registry.Execute(frame); err != nil {
			// An error here indicates an unrecoverable error, teardown transport.
			logger().Error("frugal: closing transport due to unrecoverable error processing frame: ", err)
			f.close(err)
			return
		}
	}
}

func (f *fAdapterTransport) readFrame(framedTransport *TFramedTransport) ([]byte, error) {
	_, err := framedTransport.Read([]byte{})
	if err != nil {
		return nil, err
	}
	buff := make([]byte, framedTransport.RemainingBytes())
	_, err = io.ReadFull(framedTransport, buff)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

// IsOpen returns true if the transport is open, false otherwise.
func (f *fAdapterTransport) IsOpen() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.isOpen && f.transport.IsOpen()
}

// Close closes the transport.
func (f *fAdapterTransport) Close() error {
	return f.close(nil)
}

func (f *fAdapterTransport) close(cause error) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.isOpen {
		return thrift.NewTTransportException(thrift.NOT_OPEN, "Transport not open")
	}

	f.closeSignal <- struct{}{}
	if err := f.transport.Close(); err != nil {
		// Close failed, drain close signal.
		select {
		case <-f.closeSignal:
		default:
		}
		return err
	}

	select {
	case f.closeChan <- cause:
	default:
	}
	close(f.closeChan)

	if cause == nil {
		logger().Debug("frugal: transport closed")
	} else {
		logger().Debugf("frugal: transport closed with cause: %s", cause)
	}

	// Signal transport monitor of close.
	select {
	case f.monitorCloseSignal <- cause:
	default:
	}

	f.isOpen = false
	return nil
}

// Request transmits the given data and waits for a response.
// Implementations of send should be threadsafe and respect the timeout
// present on the context.
func (f *fAdapterTransport) Request(ctx FContext, oneway bool, payload []byte) ([]byte, error) {
	resultC := make(chan []byte, 1)
	errorC := make(chan error, 1)

	if !oneway {
		f.registry.Register(ctx, resultC)
		defer f.registry.Unregister(ctx)
	}

	go f.send(payload, resultC, errorC)

	select{
	case result := <-resultC:
		return result, nil
	case err := <-errorC:
		return nil, err
	case <-time.After(ctx.Timeout()):
		return nil, thrift.NewTTransportException(thrift.TIMED_OUT, "frugal: request timed out")
	}
}

func (f *fAdapterTransport) send(payload []byte, resultC chan []byte, errorC chan error) {
	if _, err := f.transport.Write(payload); err != nil {
		errorC <- err
		return
	}
	if err := f.transport.Flush(); err != nil {
		errorC <- err
		return
	}
}

// GetRequestSizeLimit returns the maximum number of bytes that can be
// transmitted. Returns a non-positive number to indicate an unbounded
// allowable size.
func (f *fAdapterTransport) GetRequestSizeLimit() uint {
	return 0
}

// AssignOpID sets the op ID on an FContext.
func (f *fAdapterTransport) AssignOpID(ctx FContext) error {
	return f.registry.AssignOpID(ctx)
}

// SetMonitor starts a monitor that can watch the health of, and reopen,
// the transport.
func (f *fAdapterTransport) SetMonitor(monitor FTransportMonitor) {
	// Stop the previous monitor, if any.
	select {
	case f.monitorCloseSignal <- nil:
	default:
	}

	// Start the new monitor.
	monitorClosedSignal := make(chan error, 1)
	runner := &monitorRunner{
		monitor:       monitor,
		transport:     f,
		closedChannel: monitorClosedSignal,
	}
	f.monitorCloseSignal = monitorClosedSignal
	go runner.run()
}

// Closed channel receives the cause of an FTransport close (nil if clean
// close).
func (f *fAdapterTransport) Closed() <-chan error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.closeChan
}
