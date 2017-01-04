part of frugal.src.frugal;

/// Generic extension of [TError] used for frugal-specific errors.
class FError extends TError {
  /// Create an [FError] with the unknown type and empty message.
  FError() : super(0, "");

  /// Create an [FError] with the unknown type and the given message.
  FError.withMessage(String message) : super(0, message);

  /// Create an [FError] with the given type and message.
  FError.withType(int type, String message) : super(type, message);
}

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

/// Indicates a message was too large for a transport to handle.
class FMessageSizeError extends TTransportError {
  /// Create an [FMessageSizeError] with the given type and message.
  FMessageSizeError(int type, String message) : super(type, message);

  /// Create an [FMessageSizeError] indicating the request is too large.
  FMessageSizeError.request(
      {String message: "request was too large for the transport"})
      : super(FTransportError.REQUEST_TOO_LARGE, message);

  /// Create an [FMessageSizeError] indicating the response is too large.
  FMessageSizeError.response(
      {String message: "response was too large for the transport"})
      : super(FTransportError.RESPONSE_TOO_LARGE, message);
}

/// Indicates a frugal request timed out.
class FTimeoutError extends FError {
  /// Create an [FError] with the unknown type and empty message.
  FTimeoutError() : super();

  /// Create an [FError] with the unknown type and the given message.
  FTimeoutError.withMessage(String message) : super.withMessage(message);
}

/// Indicates a problem with an [FProtocol].
class FProtocolError extends TProtocolError {
  /// Create an [FError] with the optionally given type and message.
  FProtocolError([int type = TProtocolErrorType.UNKNOWN, String message = ""])
      : super(type, message);
}
