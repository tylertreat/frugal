part of frugal.src.frugal;

/// An [FTransport] that makes frugal requests via HTTP.
class FHttpTransport extends FTransport {
  /// HTTP status code for an unauthorized reqeuest.
  static const int UNAUTHORIZED = 401;

  /// HTTP status code for requesting too much data.
  static const int REQUEST_ENTITY_TOO_LARGE = 413;

  final Logger _log = new Logger('FHttpTransport');

  /// Client used by the transport to make HTTP requests.
  final wt.Client client;

  /// URI of the frugal HTTP server.
  final Uri uri;

  /// Limits the size of responses from the server.
  /// No limit will be enforced if set to a non-positive value (i.e. <1).
  final int responseSizeLimit;

  Map<String, String> _headers;

  /// Create an [FHttpTransport] instance with the given w_transport [Client],
  /// uri, and optional size restrictions, headers, and [FRegistry].
  ///
  /// If specifying headers, note that the
  ///   * content-type
  ///   * content-transfer-encoding
  ///   * accept
  ///   * x-frugal-payload-limit
  /// headers will be overwritten.
  FHttpTransport(this.client, this.uri,
      {int requestSizeLimit: 0,
      this.responseSizeLimit: 0,
      Map<String, String> additionalHeaders,
      FRegistry registry})
      : super(registry: registry, requestSizeLimit: requestSizeLimit) {
    _headers = additionalHeaders ?? {};
    _headers['content-type'] = 'application/x-frugal';
    _headers['content-transfer-encoding'] = 'base64';
    _headers['accept'] = 'application/x-frugal';
    if (responseSizeLimit > 0) {
      _headers['x-frugal-payload-limit'] = responseSizeLimit.toString();
    }
  }

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
      ..headers = _headers
      ..uri = uri
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
