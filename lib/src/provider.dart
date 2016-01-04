part of frugal;

/// Provider for Frugal Scopes
class ScopeProvider {
  final FScopeTransportFactory fTransportFactory;
  final TTransportFactory tTransportFactory;
  final TProtocolFactory tProtocolFactory;

  ScopeProvider(this.fTransportFactory, this.tTransportFactory, this.tProtocolFactory);

  ScopeTransportWithProtocol newTransportProtocol () {
    var tr = fTransportFactory.getTransport();
    if (tTransportFactory != null) {
      tr.applyProxy(tTransportFactory);
    }
//    var pr = tProtocolFactory.getProtocol(tr.thriftTransport());
    var pr = tProtocolFactory.getProtocol(tr);
    return new ScopeTransportWithProtocol(tr, pr);
  }
}

class ScopeTransportWithProtocol {
  final FScopeTransport fTransport;
  final TProtocol tProtocol;
  ScopeTransportWithProtocol(this.fTransport, this.tProtocol);
}