part of frugal;

/// Factory for FProtocol
abstract class FProtocolFactory {
  FProtocol getProtocol(TTransport transport);
}
