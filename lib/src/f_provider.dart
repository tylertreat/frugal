part of frugal;

/// Provider for Frugal Scopes
class FScopeProvider {
  final FScopeTransportFactory fTransportFactory;
  final FProtocolFactory fProtocolFactory;

  FScopeProvider(this.fTransportFactory, this.fProtocolFactory);

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

class FServiceProvider {
  final FTransport fTransport;
  final FProtocolFactory fProtocolFactory;

  FServiceProvider(this.fTransport, this.fProtocolFactory);
}
