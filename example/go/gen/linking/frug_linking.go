package linking

import (
	"fmt"
	"log"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/lib/go"
)

const (
	topicBase = "linking"
	delimiter = "."
)

type LinkingPubSub interface {
	UpdateAtoms(*AtomUpdateRequest) error
	GetCurrentAtoms(*GetCurrentAtomsRequest) error
}

type LinkingPublisher struct {
	ClientProvider map[string]*frugal.Client
	SeqId          int32
}

func NewLinkingPublisher(t frugal.TransportFactory, f thrift.TTransportFactory,
	p thrift.TProtocolFactory) *LinkingPublisher {

	return &LinkingPublisher{
		ClientProvider: newLinkingClientProvider(t, f, p),
		SeqId:          0,
	}
}

func (l *LinkingPublisher) UpdateAtoms(req *AtomUpdateRequest) error {
	op := "updateAtoms"
	topic := getLinkingPubSubTopic(op)
	client, ok := l.ClientProvider[topic]
	if !ok {
		return thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+op)
	}
	oprot := client.Protocol
	l.SeqId++
	if err := oprot.WriteMessageBegin(op, thrift.CALL, l.SeqId); err != nil {
		return err
	}
	args := UpdateAtomsArgs{
		Req: req,
	}
	if err := args.Write(oprot); err != nil {
		return err
	}
	if err := oprot.WriteMessageEnd(); err != nil {
		return err
	}
	return oprot.Flush()
}

func (l *LinkingPublisher) GetCurrentAtoms(req *GetCurrentAtomsRequest) error {
	op := "getCurrentAtoms"
	topic := getLinkingPubSubTopic(op)
	client, ok := l.ClientProvider[topic]
	if !ok {
		return thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+op)
	}
	oprot := client.Protocol
	l.SeqId++
	if err := oprot.WriteMessageBegin(op, thrift.CALL, l.SeqId); err != nil {
		return err
	}
	args := GetCurrentAtomsArgs{
		Req: req,
	}
	if err := args.Write(oprot); err != nil {
		return err
	}
	if err := oprot.WriteMessageEnd(); err != nil {
		return err
	}
	return oprot.Flush()
}

func newLinkingClientProvider(t frugal.TransportFactory, f thrift.TTransportFactory,
	p thrift.TProtocolFactory) map[string]*frugal.Client {

	provider := make(map[string]*frugal.Client)

	topic := getLinkingPubSubTopic("updateAtoms")
	transport := t.GetTransport(topic)
	if f != nil {
		transport.ApplyProxy(f)
	}
	provider[topic] = &frugal.Client{
		Protocol:  p.GetProtocol(transport.ThriftTransport()),
		Transport: transport,
	}

	topic = getLinkingPubSubTopic("getCurrentAtoms")
	transport = t.GetTransport(topic)
	if f != nil {
		transport.ApplyProxy(f)
	}
	provider[topic] = &frugal.Client{
		Protocol:  p.GetProtocol(transport.ThriftTransport()),
		Transport: transport,
	}

	return provider
}

func getLinkingPubSubTopic(op string) string {
	return fmt.Sprintf("%s%s%s", topicBase, delimiter, op)
}

type LinkingSubscriber struct {
	Handler        LinkingPubSub
	ClientProvider map[string]*frugal.Client
}

func NewLinkingSubscriber(handler LinkingPubSub, t frugal.TransportFactory,
	f thrift.TTransportFactory, p thrift.TProtocolFactory) *LinkingSubscriber {

	return &LinkingSubscriber{
		Handler:        handler,
		ClientProvider: newLinkingClientProvider(t, f, p),
	}
}

func (l *LinkingSubscriber) SubscribeUpdateAtoms() error {
	op := "updateAtoms"
	topic := getLinkingPubSubTopic(op)
	client, ok := l.ClientProvider[topic]
	if !ok {
		return thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+op)
	}

	if err := client.Transport.Subscribe(); err != nil {
		return err
	}

	go func() {
		for {
			received, err := l.recvUpdateAtoms(op, client.Protocol)
			if err != nil {
				// TODO: On what errors do we bail?
				log.Println("linking: error receiving:", err)
			}
			l.Handler.UpdateAtoms(received)
		}
	}()

	return nil
}

func (l *LinkingSubscriber) UnsubscribeUpdateAtoms() error {
	op := "updateAtoms"
	topic := getLinkingPubSubTopic(op)
	client, ok := l.ClientProvider[topic]
	if !ok {
		return thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+op)
	}

	return client.Transport.Unsubscribe()
}

func (l *LinkingSubscriber) recvUpdateAtoms(op string, iprot thrift.TProtocol) (*AtomUpdateRequest, error) {
	name, _, _, err := iprot.ReadMessageBegin()
	if err != nil {
		return nil, thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err.Error())
	}
	if name != op {
		iprot.Skip(thrift.STRUCT)
		iprot.ReadMessageEnd()
		x9 := thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+name)
		return nil, x9
	}
	args := UpdateAtomsArgs{}
	if err := args.Read(iprot); err != nil {
		iprot.ReadMessageEnd()
		x := thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err.Error())
		return nil, x
	}

	iprot.ReadMessageEnd()
	return args.Req, nil
}

func (l *LinkingSubscriber) SubscribeGetCurrentAtoms() error {
	op := "getCurrentAtoms"
	topic := getLinkingPubSubTopic(op)
	client, ok := l.ClientProvider[topic]
	if !ok {
		return thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+op)
	}

	if err := client.Transport.Subscribe(); err != nil {
		return err
	}

	go func() {
		for {
			received, err := l.recvGetCurrentAtoms(op, client.Protocol)
			if err != nil {
				// TODO: On what errors do we bail?
				log.Println("linking: error receiving:", err)
			}
			l.Handler.GetCurrentAtoms(received)
		}
	}()

	return nil
}

func (l *LinkingSubscriber) UnsubscribeGetCurrentAtoms() error {
	op := "getCurrentAtoms"
	topic := getLinkingPubSubTopic(op)
	client, ok := l.ClientProvider[topic]
	if !ok {
		return thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+op)
	}

	return client.Transport.Unsubscribe()
}

func (l *LinkingSubscriber) recvGetCurrentAtoms(op string, iprot thrift.TProtocol) (*GetCurrentAtomsRequest, error) {
	name, _, _, err := iprot.ReadMessageBegin()
	if err != nil {
		return nil, thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err.Error())
	}
	if name != op {
		iprot.Skip(thrift.STRUCT)
		iprot.ReadMessageEnd()
		x9 := thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+name)
		return nil, x9
	}
	args := GetCurrentAtomsArgs{}
	if err := args.Read(iprot); err != nil {
		iprot.ReadMessageEnd()
		x := thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err.Error())
		return nil, x
	}

	iprot.ReadMessageEnd()
	return args.Req, nil
}
