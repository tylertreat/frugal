part of frugal;

/// Wraps a Thrift TTransport. Used for frugal Scopes.
abstract class FScopeTransport extends TTransport {

  // setTopic sets the publish topic.
  void setTopic(string);

  // subscribe sets the subscribe topic and opens the transport.
  Future subscribe(string);
}