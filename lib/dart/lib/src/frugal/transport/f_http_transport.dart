part of frugal.frugal;

/// An [FTransport] that makes frugal requests via http.
class FHttpTransport extends FTransport {
  static const int UNAUTHORIZED = 401;
  static const int REQUEST_ENTITY_TOO_LARGE = 413;

  final Logger log = new Logger('FHttpTransport');
  final wt.Client client;
  final FHttpConfig config;

  FHttpTransport(this.client, config, {FRegistry registry})
      : super(registry: registry, requestSizeLimit: config.requestSizeLimit),
        this.config = config;

  @override
  bool get isOpen => true;

  @override
  Future open() => new Future.value();

  @override
  Future close([Error error]) => new Future.value();

  @override
  Future send(Uint8List payload) async {
    if (requestSizeLimit > 0 && payload.length > requestSizeLimit) {
      throw new FMessageSizeError.request();
    }

    // Encode request payload
    var requestBody = BASE64.encode(payload);

    // Configure the request
    wt.Request request = client.newRequest()
      ..headers = config.headers
      ..uri = config.uri
      ..body = requestBody;

    // Attempt the request
    wt.Response response;
    try {
      response = await request.post();
    } on StateError catch (ex) {
      throw new TTransportError(
          TTransportErrorType.UNKNOWN, 'Malformed request ${ex.toString()}');
    } on wt.RequestException catch (ex) {
      if (ex.response == null) {
        throw new TTransportError(TTransportErrorType.UNKNOWN, ex.message);
      }
      if (ex.response.status == UNAUTHORIZED) {
        throw new TTransportError(TTransportErrorType.UNKNOWN,
            'Frugal http request failed - unauthorized ${ex.message}');
      }
      if (ex.response.status == REQUEST_ENTITY_TOO_LARGE) {
        throw new FMessageSizeError.response();
      }
      throw new TTransportError(TTransportErrorType.UNKNOWN, ex.message);
    }

    // Attempt to decode the response payload
    Uint8List data;
    try {
      data = new Uint8List.fromList(BASE64.decode(response.body.asString()));
    } on FormatException catch (_) {
      throw new TProtocolError(TProtocolErrorType.INVALID_DATA,
          'Expected a Base 64 encoded string.');
    }

    // If not enough data, throw a protocol error
    if (data.length < 4) {
      throw new TProtocolError(
          TProtocolErrorType.INVALID_DATA, 'Expected frugal data to be framed');
    }

    // If there are only 4 bytes, this is a one-way request
    if (data.length == 4) {
      var bData = new ByteData.view(data.buffer);
      if (bData.getUint32(0) != 0) {
        throw new TTransportError(
            TTransportErrorType.UNKNOWN, "invalid frame size");
      }
      return;
    }

    executeFrame(data);
  }
}

/// FHttpConfig wraps request configuration information,
/// such as server URL and request headers.
class FHttpConfig {
  final Uri uri;

  /// Limits the size of Frugal frame size for requests to the server.
  /// No limit will be enforced if set to a non-positive value (i.e. <1).
  final int requestSizeLimit;

  /// Limits the size of Frugal frame size for responses from the server.
  /// No limit will be enforced if set to a non-positive value (i.e. <1).
  final int responseSizeLimit;

  Map<String, String> _headers;
  get headers => _headers;

  FHttpConfig(this.uri, Map<String, String> headers,
      {this.requestSizeLimit: 0, this.responseSizeLimit: 0}) {
    if (uri == null || !uri.hasAuthority) {
      throw new ArgumentError('Invalid uri');
    }

    _initHeaders(headers);
  }

  void _initHeaders(Map<String, String> initial) {
    var h = {};

    if (initial != null) {
      h.addAll(initial);
    }

    h['content-type'] = 'application/x-frugal';
    h['content-transfer-encoding'] = 'base64';
    h['accept'] = 'application/x-frugal';

    if (responseSizeLimit > 0) {
      h['x-frugal-payload-limit'] = responseSizeLimit.toString();
    }

    _headers = new Map.unmodifiable(h);
  }
}
