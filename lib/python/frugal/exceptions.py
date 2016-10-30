from thrift.Thrift import TApplicationException

class FException(Exception):

    def __init__(self, message=None):
        super(FException, self).__init__(message)


class FrugalVersionException(FException):

    def __init__(self, message=None):
        super(FrugalVersionException, self).__init__(message)


class FContextHeaderException(FException):

    def __init__(self, message=None):
        super(FContextHeaderException, self).__init__(message)


class FProtocolException(FException):

    UNKNOWN = 0
    INVALID_DATA = 1
    BAD_VERSION = 2

    def __init__(self, type=UNKNOWN, message=None):
        super(FProtocolException, self).__init__(message)
        self.type = type


class FExecuteCallbackNotSet(FException):

    def __init__(self, message=None):
        super(FExecuteCallbackNotSet, self).__init__(message)


class FMessageSizeException(FException):

    def __init__(self, message=None):
        super(FMessageSizeException, self).__init__(message)


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


class FOperationIdNotFoundException(FException):

    def __init__(self, message=None):
        super(FOperationIdNotFoundException, self).__init__(message)
