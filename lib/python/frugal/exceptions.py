from thrift.Thrift import TException
from thrift.protocol.TProtocol import TProtocolException


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
