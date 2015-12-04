part of frugal;

int _globalOpId = 0;

/// Context for Frugal message
class Context {
  static const _cid = "_cid";
  static const _opid= "_opid";
  Map<String, String> _requestHeaders;
  Map<String, String> _responseHeaders;

  Context({String correlationId: _generateCorrelationId()}) {
    _requestHeaders = {
      _cid: correlationId,
      _opid: _nextOpId(),
    };
    _responseHeaders = {};
  }

  Context.withRequestHeaders(Map<String, String> headers) {
    if (headers[_cid] == "") {
      headers[_cid] = _generateCorrelationId();
    }
    if (headers[_opid] = "") {
      headers[_opid] = _nextOpId();
    }
    _requestHeaders = headers;
    _responseHeaders = {};
  }

  /// Correlation Id for the context
  String correlationId() => _requestHeaders[_cid];

  /// Operation id for the context
  int opId() {
    var opIdStr = _requestHeaders[_opid];
    return int.parse(opIdStr);
  }

  /// Add a request header to the context for the given name
  void addRequestHeader(String name, value) {
    _requestHeaders[name] = value;
  }

  /// Get the named request header
  String requestHeader(String name) {
    return _requestHeaders[name];
  }

  /// Get requests headers map
  Map<String, String> requestHeaders() {
    return UnmodifiableMapView(_requestHeaders);
  }

  /// Add a response header to the context for the given name
  void addResponseHeader(String name, value) {
    _responseHeaders[name] = value;
  }

  /// Get the named response header
  String responseHeader(String name) {
    return _requestHeaders[name];
  }

  /// Get response headers map
  Map<String, String> responseHeaders() {
    return UnmodifiableMapView(_responseHeaders);
  }

  static String _generateCorrelationId() => new Uuid().v4().toString().replaceAll('-', '');

  static String _nextOpId(){
    var id = _globalOpId.toString();
    _globalOpId++;
    return id;
  }
}