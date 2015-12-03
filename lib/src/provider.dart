part of frugal;

class Provider {
  final FScopeTransportFactory fTransportFactory;
  final TTransportFactory tTransportFactory;
  final TProtocolFactory tProtocolFactory;

  Provider(this.fTransportFactory, this.tTransportFactory, this.tProtocolFactory);

  TransportWithProtocol newTransportProtocol () {
    var tr = fTransportFactory.getTransport();
    if (tTransportFactory != null) {
      tr.applyProxy(tTransportFactory);
    }
    var pr = tProtocolFactory.getProtocol(tr.thriftTransport());
    return new TransportWithProtocol(tr, pr);
  }
}

class TransportWithProtocol {
  final FScopeTransport fTransport;
  final TProtocol tProtocol;
  TransportWithProtocol(this.fTransport, this.tProtocol);
}