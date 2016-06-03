part of frugal;

/// FHttpClientTransport is a client FTransport that makes frugal requests via http.
class FHttpClientTransport extends FTransport {
  final Logger log = new Logger('FHttpTransport');
  final List<int> _writeBuffer = [];
  final http.Client httpClient;
  final FHttpConfig config;
  FRegistry _registry;

  FHttpClientTransport(this.httpClient, this.config) {}

  @override
  bool get isOpen => _registry != null;

  @override
  Future open() => new Future.value();

  @override
  Future close() => new Future.value();

  @override
  void setRegistry(FRegistry registry) {
    if (registry == null) {
      throw new FError.withMessage("registry cannot be null");
    }
    if (_registry != null) return;
    _registry = registry;
  }

  @override
  void register(FContext ctx, FAsyncCallback callback) {
    if (_registry == null) {
      throw new FError.withMessage("transport registry not set");
    }
    _registry.register(ctx, callback);
  }

  @override
  void unregister(FContext ctx) {
    if (_registry == null) {
      throw new FError.withMessage("frugal: transport registry not set");
    }
    _registry.unregister(ctx);
  }

  @override
  void write(Uint8List buffer, int offset, int length) {
    if (buffer == null) {
      throw new ArgumentError.notNull("buffer");
    }

    if (offset + length > buffer.length) {
      throw new ArgumentError("The range exceeds the buffer length");
    }

    _writeBuffer.addAll(buffer.sublist(offset, offset + length));
  }

  @override
  Future flush() async {
    var requestBody = BASE64.encode(_writeBuffer);
    _writeBuffer.clear();

    http.Response response;
    response = await httpClient.post(config.url,
        headers: config.headers, body: requestBody);
    if (response.statusCode >= 300) {
      throw new TTransportError(TTransportErrorType.UNKNOWN, response.body);
    }

    Uint8List data;
    try {
      data = new Uint8List.fromList(BASE64.decode(response.body));
    } on FormatException catch (_) {
      throw new TProtocolError(TProtocolErrorType.INVALID_DATA,
          "Expected a Base 64 encoded string.");
    }
    _registry.execute(data);
  }
}

/// FHttpConfig wraps request configuration information,
/// such as server URL and request headers.
class FHttpConfig {
  final Uri url;

  Map<String, String> _headers;
  get headers => _headers;

  FHttpConfig(this.url, Map<String, String> headers) {
    if (url == null || !url.hasAuthority) {
      throw new ArgumentError("Invalid url");
    }

    _initHeaders(headers);
  }

  void _initHeaders(Map<String, String> initial) {
    var h = {};

    if (initial != null) {
      h.addAll(initial);
    }

    h['Content-Type'] = 'application/x-frugal';
    h['Accept'] = 'application/x-frugal';

    _headers = new Map.unmodifiable(h);
  }
}
