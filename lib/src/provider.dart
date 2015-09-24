part of frugal;

class Provider {
  final TransportFactory transportFactory;
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
  final Transport transport;
  final TProtocol protocol;
  TransportWithProtocol(this.transport, this.protocol);
}