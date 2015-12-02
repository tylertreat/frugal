part of frugal;

abstract class FProtocol extends TProtocol {
  FProtocol(TTransport transport)
    : super(transport);

  void writeRequestHeader(Context);
  Context readRequestHeader();

  void writeResponseHeader(Context);
  void readResponseHeader(Context);
}
