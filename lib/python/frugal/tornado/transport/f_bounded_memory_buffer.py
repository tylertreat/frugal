from frugal.exceptions import FMessageSizeException
from thrift.transport.TTransport import TMemoryBuffer

_NATS_MAX_MESSAGE_SIZE = 1024 * 1024


class FBoundedMemoryBuffer(object, TMemoryBuffer):

    def __init__(self, value=None):
        super(FBoundedMemoryBuffer, self).__init__(value)

    def isOpen(self):
        return self._buffer.isOpen()

    def open(self):
        self._buffer.open()

    def close(self):
        self._buffer.close()

    def read(self, length):
        return self._buffer.read(length)

    def write(self, buf):
        if len(self._buffer.getvalue()) + len(buf.getvalue()) > _NATS_MAX_MESSAGE_SIZE:
            self._buffer = TMemoryBuffer(_NATS_MAX_MESSAGE_SIZE)
            raise FMessageSizeException("Buffer size reached {}".format(_NATS_MAX_MESSAGE_SIZE))
        self._buffer.write(buf)

    def length(self):
        return len(self._buffer.getvalue())

    def getvalue(self):
        return self._buffer.getvalue()
