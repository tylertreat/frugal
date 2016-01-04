part of frugal;

/// Wraps a Thrift TTransport. Used for frugal Scopes.
abstract class FScopeTransport extends TTransport {

  // LockTopic sets the publish topic and locks the transport for exclusive
  // access.
  void setTopic(string);// error

  // Subscribe sets the subscribe topic and opens the transport.
  Future subscribe(string);// error
}