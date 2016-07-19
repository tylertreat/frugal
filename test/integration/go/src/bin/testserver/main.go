package main

import (
	"flag"
	"log"

	"github.com/Workiva/frugal/test/integration/go/common"
)

var host = flag.String("host", "localhost", "Host to connect")
var port = flag.Int64("port", 9090, "Port number to connect")
var transport = flag.String("transport", "buffered", "Transport: buffered, framed, http, zlib")
var protocol = flag.String("protocol", "binary", "Protocol: binary, compact, json")

func main() {
	flag.Parse()
	server, err := common.StartServer(*host, *port, *transport, *protocol, common.PrintingHandler)
	if err != nil {
		log.Fatalln("Unable to start server: ", err)
	}
	server.Serve()
}
