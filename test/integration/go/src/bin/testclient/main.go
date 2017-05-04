package main

import (
	"flag"
	"log"

	"github.com/Workiva/frugal/test/integration/go/common"
)

var host = flag.String("host", "localhost", "Host to connect")
var port = flag.Int64("port", 9090, "Port number to connect")
var transport = flag.String("transport", "stateless", "Transport: stateless, http")
var protocol = flag.String("protocol", "binary", "Protocol: binary, compact, json")

func main() {
	flag.Parse()
	pubSub := make(chan bool)
	sent := make(chan bool)
	clientMiddlewareCalled := make(chan bool, 1)
	client, err := common.StartClient(*host, *port, *transport, *protocol, pubSub, sent, clientMiddlewareCalled)
	if err != nil {
		log.Fatal("Unable to start client: ", err)
	}

	common.CallEverything(client)

	select {
	case <-clientMiddlewareCalled:
	default:
		log.Fatal("Client middleware not invoked")
	}

	close(pubSub)
	<-sent
}

