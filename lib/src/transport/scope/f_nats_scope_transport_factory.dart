part of frugal;

/// Implementation of FScopeTransportFactory backed by FNatsScopeTransport
class FNatsScopeTransportFactory implements FTransportFactory {
  Nats client;
  FNatsScopeTransportFactory(this.client);

  FScopeTransport getTransport() => new FNatsScopeTransport(this.client);
}