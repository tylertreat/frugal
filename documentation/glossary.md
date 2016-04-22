# Glossary

This describes at a high level some of the concepts found in Frugal. Most
components in Frugal are prefixed with "F", i.e. "FTransport", in order to
differentiate from Thrift, which prefixes things with "T". Components marked
with an asterisk are internal details of Frugal and not something a user
interacts with directly but are documented for posterity. As a result, some
internal components may vary between language implementations.

## FAsyncCallback*

// TODO

## FContext

FContext is the context for a Frugal message. Every RPC has an FContext, which
can be used to set request headers, response headers, and the request timeout.
The default timeout is one minute. An FContext is also sent with every publish
message which is then received by subscribers.

In addition to headers, the FContext also contains a correlation ID which can
be used for distributed tracing purposes. A random correlation ID is generated
for each FContext if one is not provided.

FContext also plays a key role in Frugal's multiplexing support. A unique,
per-request operation ID is set on every FContext before a request is made.
This operation ID is sent in the request and included in the response, which is
then used to correlate a response to a request. The operation ID is an internal
implementation detail and is not exposed to the user.

## FProcessor

FProcessor is Frugal's equivalent of Thrift's TProcessor. It's a generic object
which operates upon an input stream and writes to an output stream.
Specifically, an FProcessor is provided to an FServer in order to wire up a
service handler to process requests.

## FProtocol

FProtocol is Frugal's equivalent of Thrift's TProtocol. It defines the
serialization protocol used for messages, such as JSON, binary, etc. FProtocol
actually extends TProtocol and adds support for serializing FContext. In
practice, FProtocol simply wraps a TProtocol and uses Thrift's built-in
serialization. FContext is encoded before the TProtocol serialization of the
message using a simple binary protocol. See the
[protocol documentation](protocol.md) for more details.

## FProtocolFactory

// TODO

## FScopeProvider

FScopeProvider produces FScopeTransports and FProtocols for use by pub/sub
scopes. It does this by wrapping an FScopeTransportFactory and
FProtocolFactory.

## FScopeTransport

FScopeTransport extends Thrift's TTransport and is used exclusively for pub/sub
scopes. Subscribers use an FScopeTransport to subscribe to a pub/sub topic.
Publishers use it to publish to a topic.

## FScopeTransportFactory

// TODO

## FServer

// TODO

## FSubscription

// TODO

## FRegistry*

// TODO

## FTransport

// TODO

## FTransportFactory

// TODO

## FTransportMonitor

// TODO
