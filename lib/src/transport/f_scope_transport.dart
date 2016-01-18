part of frugal;

/// Wraps a Thrift TTransport. Used for frugal Scopes.
abstract class FScopeTransport extends TTransport {

  /// set the publish topic.
  void setTopic(string);

  /// set the subscribe topic and opens the transport.
  Future subscribe(string);
}