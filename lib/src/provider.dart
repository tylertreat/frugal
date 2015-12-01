part of frugal;

class Provider {
  final FTransportFactory transportFactory;
  final TTransportFactory thriftTransportFactory;
  final TProtocolFactory protocolFactory;

  Provider(this.transportFactory, this.thriftTransportFactory, this.protocolFactory);

  TransportWithProtocol newTransportProtocol () {
    var tr = transportFactory.getTransport();
    if (thriftTransportFactory != null) {
      tr.applyProxy(thriftTransportFactory);
    }
    var pr = protocolFactory.getProtocol(tr.thriftTransport());
    return new TransportWithProtocol(tr, pr);
  }
}

class TransportWithProtocol {
  final FTransport transport;
  final TProtocol protocol;
  TransportWithProtocol(this.transport, this.protocol);
}