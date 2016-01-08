part of frugal;

/// Provider for Frugal Scopes
class FProvider {
  final FScopeTransportFactory fTransportFactory;
  final FProtocolFactory fProtocolFactory;

  FProvider(this.fTransportFactory, this.fProtocolFactory);

  FScopeTransportWithProtocol newTransportProtocol () {
    FScopeTransport transport = fTransportFactory.getTransport();
    FProtocol protocol = fProtocolFactory.getProtocol(transport);
    return new FScopeTransportWithProtocol(transport, protocol);
  }
}

class FScopeTransportWithProtocol {
  final FScopeTransport fTransport;
  final FProtocol fProtocol;
  FScopeTransportWithProtocol(this.fTransport, this.fProtocol);
}