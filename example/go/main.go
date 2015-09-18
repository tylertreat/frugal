package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/lib/go"
	"github.com/nats-io/nats"

	"github.com/Workiva/frugal/example/go/gen-go/event"
)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	flag.Usage = Usage
	var (
		server   = flag.Bool("server", false, "Run server")
		protocol = flag.String("P", "binary", "Specify the protocol (binary, compact, json, simplejson)")
		framed   = flag.Bool("framed", false, "Use framed transport")
		buffered = flag.Bool("buffered", false, "Use buffered transport")
		addr     = flag.String("addr", nats.DefaultURL, "NATS address")
		secure   = flag.Bool("secure", false, "Use tls secure transport")
	)
	flag.Parse()

	natsOptions := nats.DefaultOptions
	natsOptions.Servers = []string{*addr}
	natsOptions.Secure = *secure
	conn, err := natsOptions.Connect()
	if err != nil {
		panic(err)
	}

	var protocolFactory thrift.TProtocolFactory
	switch *protocol {
	case "compact":
		protocolFactory = thrift.NewTCompactProtocolFactory()
	case "simplejson":
		protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
	case "json":
		protocolFactory = thrift.NewTJSONProtocolFactory()
	case "binary", "":
		protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	default:
		fmt.Fprint(os.Stderr, "Invalid protocol specified", protocol, "\n")
		Usage()
		os.Exit(1)
	}

	var transportFactory thrift.TTransportFactory
	if *buffered {
		transportFactory = thrift.NewTBufferedTransportFactory(8192)
	} else {
		transportFactory = thrift.NewTTransportFactory()
	}

	if *framed {
		transportFactory = thrift.NewTFramedTransportFactory(transportFactory)
	}

	if *server {
		runSubscriber(conn, protocolFactory, transportFactory)
	} else {
		runPublisher(conn, protocolFactory, transportFactory)
	}
}

func runSubscriber(conn *nats.Conn, protocolFactory thrift.TProtocolFactory, transportFactory thrift.TTransportFactory) {
	factory := frugal.NewNATSTransportFactory(conn)
	subscriber := event.NewEventsSubscriber(factory, transportFactory, protocolFactory)
	_, err := subscriber.SubscribeEventCreated(func(e *event.Event) {
		fmt.Printf("received %+v\n", e)
	})
	if err != nil {
		panic(err)
	}
	ch := make(chan bool)
	log.Println("Subscriber started...")
	<-ch
}

func runPublisher(conn *nats.Conn, protocolFactory thrift.TProtocolFactory, transportFactory thrift.TTransportFactory) {
	factory := frugal.NewNATSTransportFactory(conn)
	publisher := event.NewEventsPublisher(factory, transportFactory, protocolFactory)
	event := &event.Event{ID: 42, Message: "hello, world!"}
	if err := publisher.PublishEventCreated(event); err != nil {
		panic(err)
	}
	fmt.Println("EventCreated()")
	time.Sleep(time.Second)
}
