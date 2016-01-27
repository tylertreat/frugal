package integration

import (
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

	natsOptions := nats.DefaultOptions
	natsOptions.Servers = []string{nats.DefaultURL}
	natsOptions.Secure = false // TODO: Test with TLS enabled
	conn, err := natsOptions.Connect()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for pf, protocolFactory := range protocolFactories {
		for tf, transportFactory := range transportFactories {
			name := pf + " " + tf
			PublishSubscribe(t, protocolFactory, transportFactory, conn, name)
		}
	}
}
