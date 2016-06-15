package frugal

import (
	"bytes"
	"encoding/binary"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats"
)

const defaultWorkQueueLen = 64

// FStatelessNatsServer implements FServer by using NATS as the underlying
// transport. Clients must connect with the transport created by
// NewStatelessNatsTTransport.
type FStatelessNatsServer struct {
	conn               *nats.Conn
	processor          FProcessor
	inputProtoFactory  *FProtocolFactory
	outputProtoFactory *FProtocolFactory
	subject            string
	queue              string
	workerCount        uint
	workC              chan *frameWrapper
	quit               chan struct{}
	waterMu            sync.RWMutex
	highWatermark      time.Duration
}

// NewFStatelessNatsServer creates a new FStatelessNatsServer which receives
// requests on the given subject and queue. Pass an empty string for the queue
// to not join a queue group. The worker count controls how many goroutines to
// use to process requests. This uses a default request queue length of 64.
// Clients must connect with the transport created by
// NewStatelessNatsTTransport.
func NewFStatelessNatsServer(
	conn *nats.Conn,
	processor FProcessor,
	inputProtoFactory, outputProtoFactory *FProtocolFactory,
	subject, queue string,
	workerCount uint) FServer {

	return &FStatelessNatsServer{
		conn:               conn,
		processor:          processor,
		subject:            subject,
		queue:              queue,
		workerCount:        workerCount,
		inputProtoFactory:  inputProtoFactory,
		outputProtoFactory: outputProtoFactory,
		workC:              make(chan *frameWrapper, defaultWorkQueueLen),
		quit:               make(chan struct{}),
		highWatermark:      defaultWatermark,
	}
}

// NewFStatelessNatsServerWithQueueLen creates a new FStatelessNatsServer which
// receives requests on the given subject and queue. Pass an empty string for
// the queue to not join a queue group. The worker count controls how many
// goroutines to use to process requests. The queue length controls how large
// the request queue is. Clients must connect with the transport created by
// NewStatelessNatsTTransport.
func NewFStatelessNatsServerWithQueueLen(
	conn *nats.Conn,
	processor FProcessor,
	inputProtoFactory, outputProtoFactory *FProtocolFactory,
	subject, queue string,
	workerCount, requestQueueLen uint) FServer {

	return &FStatelessNatsServer{
		conn:               conn,
		processor:          processor,
		subject:            subject,
		queue:              queue,
		workerCount:        workerCount,
		inputProtoFactory:  inputProtoFactory,
		outputProtoFactory: outputProtoFactory,
		workC:              make(chan *frameWrapper, requestQueueLen),
		quit:               make(chan struct{}),
		highWatermark:      defaultWatermark,
	}
}

// Serve starts the server.
func (f *FStatelessNatsServer) Serve() error {
	sub, err := f.conn.QueueSubscribe(f.subject, f.queue, f.handler)
	if err != nil {
		return err
	}

	for i := uint(0); i < f.workerCount; i++ {
		go f.worker()
	}

	log.Info("frugal: server running...")
	<-f.quit
	log.Info("frugal: server stopping...")

	sub.Unsubscribe()

	return nil
}

// Stop the server.
func (f *FStatelessNatsServer) Stop() error {
	close(f.quit)
	return nil
}

// SetHighWatermark sets the maximum amount of time a frame is allowed to await
// processing before triggering server overload logic. For now, this just
// consists of logging a warning. If not set, default is 5 seconds.
func (f *FStatelessNatsServer) SetHighWatermark(watermark time.Duration) {
	f.waterMu.Lock()
	f.highWatermark = watermark
	f.waterMu.Unlock()
}

// handler is invoked when a request is received. The request is placed on the
// work channel which is processed by a worker goroutine.
func (f *FStatelessNatsServer) handler(msg *nats.Msg) {
	if msg.Reply == "" {
		log.Warn("frugal: discarding invalid NATS request (no reply)")
		return
	}
	select {
	case f.workC <- &frameWrapper{frameBytes: msg.Data, timestamp: time.Now(), reply: msg.Reply}:
	case <-f.quit:
		return
	}
}

// worker should be called as a goroutine. It reads requests off the work
// channel and processes them.
func (f *FStatelessNatsServer) worker() {
	for {
		select {
		case <-f.quit:
			return
		case frame := <-f.workC:
			dur := time.Since(frame.timestamp)
			f.waterMu.RLock()
			if dur > f.highWatermark {
				log.Warnf("frugal: frame spent %+v in the transport buffer, your consumer might be backed up", dur)
			}
			f.waterMu.RUnlock()
			if err := f.processFrame(frame.frameBytes, frame.reply); err != nil {
				log.Errorf("frugal: error processing frame: %s", err.Error())
			}
		}
	}
}

// processFrame invokes the FProcessor and sends the response on the given
// subject.
func (f *FStatelessNatsServer) processFrame(frame []byte, reply string) error {
	// Read and process frame.
	input := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(frame[4:])} // Discard frame size
	outBuf := new(bytes.Buffer)
	output := &thrift.TMemoryBuffer{Buffer: outBuf}
	if err := f.processor.Process(f.inputProtoFactory.GetProtocol(input), f.outputProtoFactory.GetProtocol(output)); err != nil {
		return err
	}

	if outBuf.Len() == 0 {
		return nil
	}

	if outBuf.Len()+4 > natsMaxMessageSize {
		// QUESTION: What should happen here? With the existing NATS transport,
		// if the server attempts to send a response exceeding the limit it
		// sends a RESPONSE_TOO_LARGE exception back to the client instead.
		return ErrTooLarge
	}

	// Add frame size.
	response := make([]byte, outBuf.Len()+4)
	binary.BigEndian.PutUint32(response, uint32(outBuf.Len()))
	copy(response[4:], outBuf.Bytes())

	// Send response.
	return f.conn.Publish(reply, response)
}
