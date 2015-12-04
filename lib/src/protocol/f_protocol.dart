part of frugal;

/// Extension of TProtocol with the addition of headers
abstract class FProtocol extends TProtocol {
  FProtocol(TTransport transport)
    : super(transport);

  /// write the request headers on the given Context
  void writeRequestHeader(Context);
  /// read the requests headers into a new Context
  Context readRequestHeader();

  /// write the response headers on the given Context
  void writeResponseHeader(Context);
  /// read the requests headers into the given Context
  void readResponseHeader(Context);
}
