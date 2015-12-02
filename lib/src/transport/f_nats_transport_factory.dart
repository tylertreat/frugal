part of frugal;

class FNatsTransportFactory implements FTransportFactory {
  Nats client;

  FNatsTransportFactory(this.client);

  FTransport getTransport() => new FNatsTransport(this.client);
}