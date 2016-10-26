
class FException(Exception):
    """Basic Frugal exception."""
    def __init__(self, message=None):
        super(FException, self).__init__(message)


class FContextHeaderException(FException):
    """Indicates an invalid header key on an FContext"""
    def __init__(self, message=None):
        super(FContextHeaderException, self).__init__(message)


class FProtocolException(FException):
    """Indicates a problem with a protocol."""
    UNKNOWN = 0
    INVALID_DATA = 1
    BAD_VERSION = 2

    def __init__(self, type=UNKNOWN, message=None):
        super(FProtocolException, self).__init__(message)
        self.type = type


class FMessageSizeException(FException):
    """Indicates a message was too large for a transport to handle."""
    def __init__(self, message=None):
        super(FMessageSizeException, self).__init__(message)
