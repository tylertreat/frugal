from frugal.exceptions import FMessageSizeException
from thrift.transport.TTransport import TMemoryBuffer


class FBoundedMemoryBuffer(TMemoryBuffer, object):
    """An implementation of TMemoryBuffer using a bounded memory buffer.
    Writes which cause the buffer to exceed its size throw an FMessageSizeException"""

    def __init__(self, size, value=None):
        super(FBoundedMemoryBuffer, self).__init__(value)
        self._size = size

    def write(self, buf):
        """Bounded write to buffer"""
        if len(self) + len(buf) > self._size:
            self._buffer = TMemoryBuffer()
            raise FMessageSizeException("Buffer size reached {}".format(self._size))
        self._buffer.write(buf)

    def __len__(self):
        return len(self.getvalue())
