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
      throw new FError.withMessage('registry cannot be null');
    }
    if (_registry != null) return;
    _registry = registry;
  }

  @override
  void register(FContext ctx, FAsyncCallback callback) {
    if (_registry == null) {
      throw new FError.withMessage('transport registry not set');
    }
    _registry.register(ctx, callback);
  }

  @override
  void unregister(FContext ctx) {
    if (_registry == null) {
      throw new FError.withMessage('frugal: transport registry not set');
    }
    _registry.unregister(ctx);
  }

  @override
  void write(Uint8List buffer, int offset, int length) {
    if (buffer == null) {
      throw new ArgumentError.notNull('buffer');
    }

    if (offset + length > buffer.length) {
      throw new ArgumentError('The range exceeds the buffer length');
    }

    if (config.requestSizeLimit > 0 &&
        length + _writeBuffer.length > config.requestSizeLimit) {
      throw new FMessageSizeError.request();
    }

    _writeBuffer.addAll(buffer.sublist(offset, offset + length));
  }

  @override
  Future flush() async {
    // Frame the request body per frugal spec
    Uint8List bytes = new Uint8List(4 + _writeBuffer.length);
    bytes.buffer.asByteData().setUint32(0, _writeBuffer.length);
    bytes.setAll(4, _writeBuffer);
    _writeBuffer.clear();

    // Encode request
    var requestBody = BASE64.encode(bytes);

    // Make the POST request
    http.Response response;
    try {
      response = await httpClient.post(config.url,
          headers: config.headers, body: requestBody);
    } catch (e) {
      throw new TTransportError(TTransportErrorType.UNKNOWN, e.toString());
    }

    // Requested too much data or other possible request errors
    if (response.statusCode == 413) {
      throw new FMessageSizeError.response();
    } else if (response.statusCode >= 300) {
      throw new TTransportError(TTransportErrorType.UNKNOWN, response.body);
    }

    // Attempt to decode the response payload
    Uint8List data;
    try {
      data = new Uint8List.fromList(BASE64.decode(response.body));
    } on FormatException catch (_) {
      throw new TProtocolError(TProtocolErrorType.INVALID_DATA,
          'Expected a Base 64 encoded string.');
    }

    // Process the request, but drop the frame size
    _registry.execute(data.sublist(4));
  }
}

/// FHttpConfig wraps request configuration information,
/// such as server URL and request headers.
class FHttpConfig {
  final Uri url;

  /// Limits the size of Frugal frame size for requests to the server.
  /// No limit will be enforced if set to a non-positive value (i.e. <1).
  final int requestSizeLimit;

  /// Limits the size of Frugal frame size for responses from the server.
  /// No limit will be enforced if set to a non-positive value (i.e. <1).
  final int responseSizeLimit;

  Map<String, String> _headers;
  get headers => _headers;

  FHttpConfig(this.url, Map<String, String> headers,
      {this.requestSizeLimit: 0, this.responseSizeLimit: 0}) {
    if (url == null || !url.hasAuthority) {
      throw new ArgumentError('Invalid url');
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

    if (responseSizeLimit > 0) {
      h['X-Frugal-Payload-Limit'] = responseSizeLimit.toString();
    }

    _headers = new Map.unmodifiable(h);
  }
}
