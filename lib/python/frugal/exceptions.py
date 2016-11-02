from thrift.Thrift import TException
from thrift.protocol.TProtocol import TProtocolException
from thrift.Thrift import TApplicationException


class FException(TException):
    """Basic Frugal exception."""
    def __init__(self, message=None):
        super(FException, self).__init__(message)


class FContextException(FException):
    """Indicates a problem with an FContext"""
    def __init__(self, message=None):
        super(FException, self).__init__(message=message)


class FContextHeaderException(FException):
    """Indicates an invalid header key on an FContext"""
    def __init__(self, message=None):
        super(FContextHeaderException, self).__init__(message)


class FProtocolException(TProtocolException):
    """Indicates a problem with a protocol."""
    def __init__(self, kind=TProtocolException.UNKNOWN, message=None):
        super(FProtocolException, self).__init__(type=kind, message=message)


class FMessageSizeException(FException):
    """Indicates a message was too large for a transport to handle."""
    def __init__(self, message=None):
        super(FMessageSizeException, self).__init__(message)


class FTimeoutException(FException):
    """Indicates a request took too long."""
    def __init__(self, message=None):
        super(FTimeoutException, self).__init__(message=message)


class FRateLimitException(TApplicationException):
    """
    FRateLimitException indicates that an application has exceeded a rate
    limit threshold.
    """

    RATE_LIMIT_EXCEEDED = 102

    def __init__(self, message="rate limit exceeded"):
        """
        Args:
            message: exception message to provide with the rate limit error.
            Defaults to "rate limit exceeded".
        """
        super(FRateLimitException, self).__init__(
            type=FRateLimitException.RATE_LIMIT_EXCEEDED, message=message)
