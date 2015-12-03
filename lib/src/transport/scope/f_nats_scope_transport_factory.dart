part of frugal;

/// Implementation of FScopeTransportFactory backed by FNatsScopeTransport
class FNatsScopeTransportFactory implements FScopeTransportFactory {
  Nats client;
  FNatsScopeTransportFactory(this.client);

  FScopeTransport getTransport() => new FNatsScopeTransport(this.client);
}