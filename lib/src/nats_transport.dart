part of frugal;

class NatsTransportFactory implements TransportFactory {
  Nats client;

  NatsTransportFactory(this.client);

  Transport getTransport() => new NatsTransport(this.client);
}


class NatsTransport implements Transport {
  TTransport tTransport;
  NatsThriftTransport nTransport;

  Stream get signalRead => nTransport.signalRead;
  Stream get error => nTransport.error;

  NatsTransport(Nats client) {
    var tr = new NatsThriftTransport(client);
    tTransport = tr;
    nTransport = tr;
  }

  Future subscribe(String subject) {
    nTransport.setSubject(subject);
    return nTransport.open();
  }

  Future unsubscribe() {
    return nTransport.close();
  }

  void preparePublish(String subject) {
    nTransport.setSubject(subject);
  }

  TTransport thriftTransport() => tTransport;

  Future applyProxy(TTransportFactory transportFactory) async {
    tTransport = await transportFactory.getTransport(tTransport);
  }
}
