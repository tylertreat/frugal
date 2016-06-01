part of frugal;

/// FMultiplexedTransport is a multiplexed Transport that routes frames to the
/// correct callbacks.
class FHttpTransport extends FTransport {
  final Logger log = new Logger('FHttpTransport');
  final List<int> _writeBuffer = [];
  final http.Client httpClient;
  final THttpConfig config;
  FRegistry _registry;

  FHttpTransport(this.httpClient, this.config) {}

  @override
  bool get isOpen => _transport.isOpen && _registry != null;

  @override
  Future open() => new Future.value();

  @override
  Future close() => closeWithException(null);

  // TODO: Throw error if direct read

  Future closeWithException(cause) async {
    await _transport.close();
    await _signalClose(cause);
  }

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
    Uint8List bytes = new Uint8List(4 + _writeBuffer.length);
    bytes.buffer.asByteData().setUint32(0, _writeBuffer.length);
    bytes.setAll(4, _writeBuffer);
    _writeBuffer.clear();
    var requestBody = BASE64.encode(bytes);

    // Use a sync completer to ensure that the buffer can be read immediately
    // after the read buffer is set, and avoid a race condition where another
    // response could overwrite the read buffer.
    var completer = new Completer.sync();

    httpClient
        .post(config.url, headers: config.headers, body: requestBody)
        .then((response) {
      Uint8List data;
      try {
        data = new Uint8List.fromList(BASE64.decode(response.body));
      } on FormatException catch (_) {
        throw new TProtocolError(TProtocolErrorType.INVALID_DATA,
            "Expected a Base 64 encoded string.");
      }
      completer.complete();
      _registry.execute(data.sublist(4));
    });

    return completer.future;
  }
}
