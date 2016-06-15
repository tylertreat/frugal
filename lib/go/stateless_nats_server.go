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
	sub                *nats.Subscription
	waterMu            sync.RWMutex
	highWatermark      time.Duration
}

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

func (f *FStatelessNatsServer) Serve() error {
	sub, err := f.conn.QueueSubscribe(f.subject, f.queue, f.handler)
	if err != nil {
		return err
	}
	f.sub = sub

	for i := uint(0); i < f.workerCount; i++ {
		go f.worker()
	}

	log.Info("frugal: server running...")
	<-f.quit
	log.Info("frugal: server stopping...")
	if f.conn.Status() != nats.CONNECTED {
		log.Warn("frugal: Nats is already disconnected!")
		return nil
	}
	return nil
}

func (f *FStatelessNatsServer) Stop() error {
	close(f.quit)
	return nil
}

func (f *FStatelessNatsServer) SetHighWatermark(watermark time.Duration) {
	f.waterMu.Lock()
	f.highWatermark = watermark
	f.waterMu.Unlock()
}

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
		return ErrTooLarge
	}

	// Add frame size.
	response := make([]byte, outBuf.Len()+4)
	binary.BigEndian.PutUint32(response, uint32(outBuf.Len()))
	copy(response[4:], outBuf.Bytes())

	// Send response.
	return f.conn.Publish(reply, response)
}
