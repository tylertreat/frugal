part of frugal;

int _nextOpId = 0;

class Context {
  static const _cid = "_cid";
  static const _opid= "_opid";
  Map<String, String> _requestHeaders;
  Map<String, String> _responseHeaders;

  Context(String correlationId) {
    if (correlationId == "") {
      correlationId = "foo";
    }
    _requestHeaders = {
      _cid: correlationId,
      _opid: _nextOpId.toString(),
    };
    _responseHeaders = {};
    _nextOpId++;
  }

  String correlationId() => _requestHeaders[_cid];

  int opId() {
    var opIdStr = _requestHeaders[_opid];
    return int.parse(opIdStr);
  }

  void addRequestHeader(String name, value) {
    _requestHeaders[name] = value;
  }

  String requestHeader(String name) {
    return _requestHeaders[name];
  }

  Map<String, String> requestHeaders() {
    return _requestHeaders;
  }

  void addResponseHeader(String name, value) {
    _responseHeaders[name] = value;
  }

  String responseHeader(String name) {
    return _requestHeaders[name];
  }

  Map<String, String> responseHeaders() {
    return _responseHeaders;
  }
}