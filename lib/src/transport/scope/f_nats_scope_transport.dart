part of frugal;

/// Implementation of FScopeTransport backed by NATS client
class FNatsScopeTransport implements FScopeTransport {
  TTransport tTransport;
  TNatsScopeTransport nTransport;

  /// Stream that signals to transport listener that data is available on
  /// the thrift transport.
  Stream get signalRead => nTransport.signalRead;
  /// Stream that signals to transport listener that there was an error.
  Stream get error => nTransport.error;

  FNatsScopeTransport(Nats client) {
    var tr = new TNatsScopeTransport(client);
    tTransport = tr;
    nTransport = tr;
  }

  /// Open the Transport to receive messages on the subscription.
  Future subscribe(String subject) {
    nTransport.setSubject(subject);
    return nTransport.open();
  }

  /// Close the Transport to stop receiving messages on the subscription.
  Future unsubscribe() {
    return nTransport.close();
  }

  /// Prepare the Transport for publishing to the given topic.
  void preparePublish(String subject) {
    nTransport.setSubject(subject);
  }

  /// Return the wrapped Thrift TTransport.
  TTransport thriftTransport() => tTransport;

  /// Wrap the underlying TTransport with the TTransport returned by the
  /// given TTransportFactory.
  Future applyProxy(TTransportFactory transportFactory) async {
    tTransport = await transportFactory.getTransport(tTransport);
  }
}
