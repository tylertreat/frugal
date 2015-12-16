package integration

import (
	"flag"
	"testing"

	"github.com/nats-io/nats"

	"git.apache.org/thrift.git/lib/go/thrift"
)

func TestPublishSubscribe(t *testing.T) {

	protocolFactories := map[string]thrift.TProtocolFactory{
		"TCompactProtocolFactory":       thrift.NewTCompactProtocolFactory(),
		"TJSONProtocolFactory":          thrift.NewTJSONProtocolFactory(),
		"TBinaryProtocolFactoryDefault": thrift.NewTBinaryProtocolFactoryDefault(),
	}
	transportFactories := map[string]thrift.TTransportFactory{
		"TBufferedTransportFactory": thrift.NewTBufferedTransportFactory(8192),
		"TTransportFactory":         thrift.NewTTransportFactory(),
	}

	tls := false // TODO: test with TLS enabled

	addr := flag.String("addr", nats.DefaultURL, "NATS address")
	secure := flag.Bool("secure", tls, "Use tls secure transport")

	natsOptions := nats.DefaultOptions
	natsOptions.Servers = []string{*addr}
	natsOptions.Secure = *secure
	conn, err := natsOptions.Connect()
	if err != nil {
		panic(err)
	}

	for pf, protocolFactory := range protocolFactories {
		for tf, transportFactory := range transportFactories {
			name := pf + " " + tf
			PublishSubscribe(t, protocolFactory, transportFactory, conn, name)
		}
	}

}
