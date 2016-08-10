// TODO: Remove this with 2.0
package frugal

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
)

const (
	queue                      = "rpc"
	defaultMaxMissedHeartbeats = 3
	defaultNumWorkers          = 5
	leftBrace                  = '\x7b' // "{"
)

type client struct {
	tr            thrift.TTransport
	stopHeartbeat chan struct{}
	heartbeat     string
}

func newInbox(prefix string) string {
	tokens := strings.Split(prefix, ".")
	tokens[len(tokens)-1] = nats.NewInbox() // Always at least 1 token
	inbox := ""
	pre := ""
	for _, token := range tokens {
		inbox += pre + token
		pre = "."
	}
	return inbox
}

// FNatsServer implements FServer by using NATS as the underlying transport.
// Clients must connect with the transport created by NewNatsServiceTTransport
// or NewFNatsTransport.
// DEPRECATED - With the next major release of frugal, stateful NATS transports
// will no longer be supported. Use the "stateless" Nats server instead with
// FNatsServerBuilder.
type FNatsServer struct {
	conn                *nats.Conn
	subjects            []string
	heartbeatSubject    string
	heartbeatInterval   time.Duration
	maxMissedHeartbeats uint
	clients             map[string]*client
	mu                  sync.Mutex
	quit                chan struct{}
	processorFactory    FProcessorFactory
	transportFactory    FTransportFactory
	protocolFactory     *FProtocolFactory
	highWatermark       time.Duration
	waterMu             sync.RWMutex
	numWorkers          uint
	workC               chan *frameWrapper
}

// NewFNatsServer creates a new FNatsServer which listens for requests on the
// given subject.
// Clients must connect with the transport created by NewNatsServiceTTransport
// or NewFNatsTransport.
// DEPRECATED - With the next major release of frugal, stateful NATS transports
// will no longer be supported. With the release of 2.0, FStatelessNatsServer
// will be renamed to FNatsServer.
func NewFNatsServer(
	conn *nats.Conn,
	subject string,
	heartbeatInterval time.Duration,
	processor FProcessor,
	transportFactory FTransportFactory,
	protocolFactory *FProtocolFactory) FServer {

	return NewFNatsServerFactory(
		conn,
		subject,
		heartbeatInterval,
		defaultMaxMissedHeartbeats,
		NewFProcessorFactory(processor),
		transportFactory,
		protocolFactory,
	)
}

// NewFNatsServerWithSubjects creates a new FNatsServer which listens for
// requests on the given subjects.
// Clients must connect with the transport created by NewNatsServiceTTransport
// or NewFNatsTransport.
// DEPRECATED - With the next major release of frugal, stateful NATS transports
// will no longer be supported. With the release of 2.0, FStatelessNatsServer
// will be renamed to FNatsServer.
func NewFNatsServerWithSubjects(
	conn *nats.Conn,
	subjects []string,
	heartbeatInterval time.Duration,
	processor FProcessor,
	transportFactory FTransportFactory,
	protocolFactory *FProtocolFactory) FServer {

	return NewFNatsServerFactoryWithSubjects(
		conn,
		subjects,
		heartbeatInterval,
		defaultMaxMissedHeartbeats,
		NewFProcessorFactory(processor),
		transportFactory,
		protocolFactory,
	)
}

// NewFNatsServerFactory creates a new FNatsServer which listens for requests
// on the given subject.
// Clients must connect with the transport created by NewNatsServiceTTransport
// or NewFNatsTransport.
// DEPRECATED - With the next major release of frugal, stateful NATS transports
// will no longer be supported. With the release of 2.0, FStatelessNatsServer
// will be renamed to FNatsServer.
func NewFNatsServerFactory(
	conn *nats.Conn,
	subject string,
	heartbeatInterval time.Duration,
	maxMissedHeartbeats uint,
	processorFactory FProcessorFactory,
	transportFactory FTransportFactory,
	protocolFactory *FProtocolFactory) FServer {

	return NewFNatsServerFactoryWithSubjects(
		conn,
		[]string{subject},
		heartbeatInterval,
		maxMissedHeartbeats,
		processorFactory,
		transportFactory,
		protocolFactory,
	)
}

// NewFNatsServerFactoryWithSubjects creates a new FNatsServer which listens
// for requests on the given subjects.
// Clients must connect with the transport created by NewNatsServiceTTransport
// or NewFNatsTransport.
// DEPRECATED - With the next major release of frugal, stateful NATS transports
// will no longer be supported. With the release of 2.0, FStatelessNatsServer
// will be renamed to FNatsServer.
func NewFNatsServerFactoryWithSubjects(
	conn *nats.Conn,
	subjects []string,
	heartbeatInterval time.Duration,
	maxMissedHeartbeats uint,
	processorFactory FProcessorFactory,
	transportFactory FTransportFactory,
	protocolFactory *FProtocolFactory) FServer {

	return NewFNatsServerWithStatelessConfig(
		conn,
		subjects,
		defaultNumWorkers,
		defaultWorkQueueLen,
		heartbeatInterval,
		maxMissedHeartbeats,
		processorFactory,
		transportFactory,
		protocolFactory)
}

// NewFNatsServerWithStatelessConfig creates a new FNatsServer which listens
// for requests on the given subjects, and has the given number of stateless
// worker goroutines and stateless request buffer queue length.
// Clients must connect with the transport created by NewNatsServiceTTransport
// or NewFNatsTransport.
// DEPRECATED - With the next major release of frugal, stateful NATS transports
// will no longer be supported. With the release of 2.0, FStatelessNatsServer
// will be renamed to FNatsServer.
func NewFNatsServerWithStatelessConfig(
	conn *nats.Conn,
	subjects []string,
	numWorkers uint,
	queueLen uint,
	heartbeatInterval time.Duration,
	maxMissedHeartbeats uint,
	processorFactory FProcessorFactory,
	transportFactory FTransportFactory,
	protocolFactory *FProtocolFactory) FServer {

	return &FNatsServer{
		conn:                conn,
		subjects:            subjects,
		numWorkers:          numWorkers,
		heartbeatSubject:    nats.NewInbox(),
		heartbeatInterval:   heartbeatInterval,
		maxMissedHeartbeats: maxMissedHeartbeats,
		clients:             make(map[string]*client),
		processorFactory:    processorFactory,
		transportFactory:    transportFactory,
		protocolFactory:     protocolFactory,
		quit:                make(chan struct{}, 1),
		highWatermark:       defaultWatermark,
		workC:               make(chan *frameWrapper, queueLen),
	}
}

// Serve starts the server.
func (n *FNatsServer) Serve() error {
	subscriptions := make([]*nats.Subscription, len(n.subjects))
	for i, subject := range n.subjects {
		sub, err := n.conn.QueueSubscribe(subject, queue, n.handleConnection)
		if err != nil {
			return err
		}
		subscriptions[i] = sub
	}

	n.conn.Flush()
	if n.isHeartbeating() {
		go n.startHeartbeat()
	}

	for i := uint(0); i < n.numWorkers; i++ {
		go n.worker()
	}

	logger().Info("frugal: server running...")
	<-n.quit
	logger().Info("frugal: server stopping...")

	for _, sub := range subscriptions {
		sub.Unsubscribe()
	}

	return nil
}

// Stop the server.
func (n *FNatsServer) Stop() error {
	close(n.quit)
	return nil
}

// SetHighWatermark sets the maximum amount of time a frame is allowed to await
// processing before triggering server overload logic. For now, this just
// consists of logging a warning. If not set, default is 5 seconds.
func (n *FNatsServer) SetHighWatermark(watermark time.Duration) {
	n.waterMu.Lock()
	n.highWatermark = watermark
	n.waterMu.Unlock()
}

// handleConnection is invoked when a remote peer is attempting to connect to
// the server.
func (n *FNatsServer) handleConnection(msg *nats.Msg) {
	if msg.Reply == "" {
		logger().Warnf("frugal: discarding invalid connect message %+v", msg)
		return
	}

	// Check to see if this is a stateless client
	if msg.Data != nil && len(msg.Data) > 0 && msg.Data[0] != leftBrace {
		select {
		case n.workC <- &frameWrapper{frameBytes: msg.Data, timestamp: time.Now(), reply: msg.Reply}:
		case <-n.quit:
			return
		}
		return
	}

	hs := &natsConnectionHandshake{}
	if err := json.Unmarshal(msg.Data, hs); err != nil {
		logger().Errorf("frugal: could not deserialize connect message %+v", msg)
		return
	}
	if hs.Version != natsV0 {
		logger().Errorf("frugal: not a supported connect version %d", hs.Version)
		return
	}
	var (
		heartbeatReply = nats.NewInbox()
		listenTo       = newInbox(msg.Reply)
		tr, err        = n.accept(listenTo, msg.Reply, heartbeatReply)
	)
	if err != nil {
		logger().Errorf("frugal: error accepting client transport: %s", err)
		return
	}

	client := &client{tr: tr, stopHeartbeat: make(chan struct{}), heartbeat: heartbeatReply}
	if n.isHeartbeating() {
		n.mu.Lock()
		n.clients[heartbeatReply] = client
		n.mu.Unlock()
	}

	// Connect message consists of "[heartbeat subject] [heartbeat reply subject] [expected interval ms]"
	connectMsg := n.heartbeatSubject + " " + heartbeatReply + " " +
		strconv.FormatInt(int64(n.heartbeatInterval/time.Millisecond), 10)
	if err := n.conn.PublishRequest(msg.Reply, listenTo, []byte(connectMsg)); err != nil {
		logger().Errorf("frugal: error publishing transport inbox: %s", err)
		tr.Close()
	} else if n.isHeartbeating() {
		go n.acceptHeartbeat(client)
	}
}

func (n *FNatsServer) remove(heartbeat string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	client, ok := n.clients[heartbeat]
	if !ok {
		return
	}
	client.tr.Close()
	close(client.stopHeartbeat)
	delete(n.clients, heartbeat)
}

func (n *FNatsServer) startHeartbeat() {
	hbTicker := time.NewTicker(n.heartbeatInterval)
	for {
		select {
		case <-hbTicker.C:
			n.mu.Lock()
			clients := len(n.clients)
			n.mu.Unlock()
			if clients == 0 {
				continue
			}
			if err := n.conn.Publish(n.heartbeatSubject, nil); err != nil {
				logger().Error("frugal: error publishing heartbeat:", err.Error())
			}
			if err := n.conn.FlushTimeout(n.heartbeatInterval * 3 / 4); err != nil {
				logger().Error("frugal: error flushing heartbeat:", err.Error())
			}
		case <-n.quit:
			return
		}
	}
}

func (n *FNatsServer) acceptHeartbeat(client *client) {
	missed := uint(0)
	recvHeartbeat := make(chan struct{}, 1)

	sub, err := n.conn.Subscribe(client.heartbeat, func(msg *nats.Msg) {
		select {
		case recvHeartbeat <- struct{}{}:
		default:
			logger().Info("frugal: FNatsServer dropped heartbeat:", client.heartbeat)
		}
	})
	if err != nil {
		logger().Error("frugal: error subscribing to heartbeat:", client.heartbeat)
		return
	}
	defer sub.Unsubscribe()

	var wait <-chan time.Time
	for {
		if n.maxMissedHeartbeats > 1 {
			wait = time.After(n.heartbeatInterval)
		} else {
			wait = time.After(n.heartbeatInterval + n.heartbeatInterval/4)
		}
		select {
		case <-wait:
			missed++
			if missed >= n.maxMissedHeartbeats {
				logger().Warnf("frugal: client heartbeat expired for heartbeat: %s", client.heartbeat)
				n.remove(client.heartbeat)
				return
			}
		case <-recvHeartbeat:
			missed = 0
		case <-client.stopHeartbeat:
			return
		case <-n.quit:
			return
		}
	}
}

func (n *FNatsServer) accept(listenTo, replyTo, heartbeat string) (FTransport, error) {
	client := newNatsServiceTTransportServer(n.conn, listenTo, replyTo)
	transport := n.transportFactory.GetTransport(client)
	processor := n.processorFactory.GetProcessor(transport)
	protocol := n.protocolFactory.GetProtocol(transport)
	transport.SetRegistry(NewServerRegistry(processor, n.protocolFactory, protocol))
	n.waterMu.RLock()
	transport.SetHighWatermark(n.highWatermark)
	n.waterMu.RUnlock()
	if err := transport.Open(); err != nil {
		return nil, err
	}

	// Cleanup heartbeat when client disconnects.
	go func() {
		select {
		case <-n.quit:
			client.Close()
		case <-transport.Closed():
		}

		n.remove(heartbeat)
	}()

	logger().Debug("frugal: client connection accepted with heartbeat:", heartbeat)
	return transport, nil
}

func (n *FNatsServer) isHeartbeating() bool {
	return n.heartbeatInterval > 0
}

// worker should be called as a goroutine. It reads requests off the work
// channel and processes them.
func (n *FNatsServer) worker() {
	for {
		select {
		case <-n.quit:
			return
		case frame := <-n.workC:
			dur := time.Since(frame.timestamp)
			n.waterMu.RLock()
			if dur > n.highWatermark {
				logger().Warnf("frugal: frame spent %+v in the transport buffer, your consumer might be backed up", dur)
			}
			n.waterMu.RUnlock()
			if err := n.processFrame(frame.frameBytes, frame.reply); err != nil {
				logger().Errorf("frugal: error processing frame: %s", err.Error())
			}
		}
	}
}

// processFrame invokes the FProcessor and sends the response on the given
// subject.
func (n *FNatsServer) processFrame(frame []byte, reply string) error {
	// Read and process frame.
	input := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(frame[4:])} // Discard frame size
	// Buffer 1MB - 4 bytes since frame size is copied directly.
	output := NewFBoundedMemoryBuffer(natsMaxMessageSize - 4)
	if err := n.processorFactory.GetProcessor(nil).Process(
		n.protocolFactory.GetProtocol(input),
		n.protocolFactory.GetProtocol(output)); err != nil {
		return err
	}

	if output.Len() == 0 {
		return nil
	}

	// Add frame size (4-byte uint32).
	response := make([]byte, output.Len()+4)
	binary.BigEndian.PutUint32(response, uint32(output.Len()))
	copy(response[4:], output.Bytes())

	// Send response.
	return n.conn.Publish(reply, response)
}
