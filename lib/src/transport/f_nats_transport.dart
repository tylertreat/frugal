part of frugal;

class FNatsTransport implements FTransport {
  TTransport tTransport;
  TNatsTransport nTransport;

  Stream get signalRead => nTransport.signalRead;
  Stream get error => nTransport.error;

  FNatsTransport(Nats client) {
    var tr = new TNatsTransport(client);
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
