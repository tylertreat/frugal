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

/// Indicates a message was too large for a transport to handle.
class FMessageSizeError extends TTransportError {
  /// Create an [FMessageSizeError] with the given type and message.
  FMessageSizeError(int type, String message) : super(type, message);

  /// Create an [FMessageSizeError] indicating the request is too large.
  FMessageSizeError.request()
      : super(FTransport.REQUEST_TOO_LARGE,
            "request was too large for the transport");

  /// Create an [FMessageSizeError] indicating the response is too large.
  FMessageSizeError.response()
      : super(FTransport.RESPONSE_TOO_LARGE,
            "response was too large for the transport");
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
