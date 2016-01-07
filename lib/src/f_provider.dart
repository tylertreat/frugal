part of frugal;

/// Provider for Frugal Scopes
class FProvider {
  final FScopeTransportFactory fTransportFactory;
  final TTransportFactory tTransportFactory;
  final TProtocolFactory tProtocolFactory;

  FProvider(this.fTransportFactory, this.tTransportFactory, this.tProtocolFactory);

  FScopeTransportWithProtocol newTransportProtocol () {
    var tr = fTransportFactory.getTransport();
    if (tTransportFactory != null) {
      tr.applyProxy(tTransportFactory);
    }
    var pr = tProtocolFactory.getProtocol(tr);
    return new FScopeTransportWithProtocol(tr, pr);
  }
}

class FScopeTransportWithProtocol {
  final FScopeTransport fTransport;
  final TProtocol tProtocol;
  FScopeTransportWithProtocol(this.fTransport, this.tProtocol);
}