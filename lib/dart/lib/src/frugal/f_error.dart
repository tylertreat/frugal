part of frugal.src.frugal;

/// Generic extension of [TError] used for frugal-specific errors.
class FError extends TError {
  FError() : super(0, "");

  FError.withMessage(String message) : super(0, message);

  FError.withType(int type, String message) : super(type, message);
}

/// Indicates a message was too large for a transport to handle.
class FMessageSizeError extends TTransportError {
  FMessageSizeError(int type, String message) : super(type, message);

  FMessageSizeError.request()
      : super(FTransport.REQUEST_TOO_LARGE,
            "request was too large for the transport");

  FMessageSizeError.response()
      : super(FTransport.RESPONSE_TOO_LARGE,
            "response was too large for the transport");
}

/// Indicates a frugal request timed out.
class FTimeoutError extends FError {
  FTimeoutError() : super();

  FTimeoutError.withMessage(String message) : super.withMessage(message);
}

/// Indicates a problem with an [FProtocol].
class FProtocolError extends TProtocolError {
  FProtocolError([int type = TProtocolErrorType.UNKNOWN, String message = ""])
      : super(type, message);
}
