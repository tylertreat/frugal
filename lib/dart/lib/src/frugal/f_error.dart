part of frugal.src.frugal;

/// Contains [TApplicationError] constants.
class FApplicationError extends TApplicationError {
  /// Indicates the response was too large for the transport.
  static const int RESPONSE_TOO_LARGE = 100;
}

/// Contains [TTransportError] constants.
class FTransportError extends TTransportError {
  /// Indicates the request was too large for the transport.
  static const int REQUEST_TOO_LARGE = 100;

  /// Indicates the response was too large for the transport.
  static const int RESPONSE_TOO_LARGE = 101;
}
