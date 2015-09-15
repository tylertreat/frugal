# Frugal

Frugal is an extension of [Apache Thrift](https://thrift.apache.org/) which provides a framework and code generation for pub/sub messaging.

## Installation

```
$ go get github.com/Workiva/frugal
```

## Usage

Define your Thrift file which contains your structs.

```thrift
// event.thrift
struct Event {
    1: i64 ID,
    2: string Message
}
```

Define your Frugal file which contains your pub/sub interface. The structs referenced here must be defined in the corresponding Thrift file, and the two files should have the same name.

```thrift
// event.frugal
namespace Events {
    EventCreated: Event
}
```

Generate the code with `frugal`. Currently, only Go is supported. The `-file` flag is the base file name shared by your Thrift and Frugal files.

```
$ frugal -gen=go -file=event
```

By default, generated code is placed in a `gen-*` directory. This code can then be used as such:

```go
// publisher.go
func main() {
    options := nats.DefaultOptions
    conn, err := options.Connect()
    if err != nil {
        panic(err)
    }

    var (
        protocolFactory  = thrift.NewTBinaryProtocolFactoryDefault()
        transportFactory = thrift.NewTTransportFactory()
        frugalFactory    = frugal.NewNATSTransportFactory(conn)
        publisher        = event.NewEventsPublisher(frugalFactory, transportFactory, protocolFactory)
    )
    
    event := &event.Event{ID: 42, Message: "Hello, World!"}
    if err := publisher.EventCreated(event); err != nil {
        panic(err)
    }
}
```

```go
// subscriber.go
func main() {
  options := nats.DefaultOptions
  conn, err := options.Connect()
  if err != nil {
      panic(err)
  }
  
  var (
      protocolFactory  = thrift.NewTBinaryProtocolFactoryDefault()
      transportFactory = thrift.NewTTransportFactory()
      frugalFactory    = frugal.NewNATSTransportFactory(conn)
      subscriber       = event.NewEventSubscriber(&eventHandler{}, frugalFactory, transportFactory, protocolFactory)
  )
  
  if err := subscriber.SubscribeEventCreated(); err != nil {
      panic(err)
  }
  
  wait := make(chan bool)
  log.Println("Subscriber started...")
  <-wait
}

type eventHandler struct {}

func (e *eventHandler) EventCreated(event *event.Event) error {
    fmt.Println("Received event:", event.Message)
    return nil
}
```
