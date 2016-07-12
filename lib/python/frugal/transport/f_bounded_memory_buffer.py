from frugal.exceptions import FMessageSizeException
from thrift.transport.TTransport import TMemoryBuffer


class FBoundedMemoryBuffer(TMemoryBuffer, object):
    """
    An implementation of TMemoryBuffer using a bounded memory buffer. Writes
    that cause the buffer to exceed its size throw an FMessageSizeException
    """

    def __init__(self, size, value=None):
        """Create an instance of FBoundedMemoryBuffer where size is the
        maximum writable length of the buffer.

           Args:
               size: integer size limit of the buffer
               value: optional data value to initialize the buffer with.
        """
        super(FBoundedMemoryBuffer, self).__init__(value)
        self._limit = size

    def write(self, buf):
        """Bounded write to buffer"""
        if len(self) + len(buf) > self._limit:
            self._buffer = TMemoryBuffer()
            raise FMessageSizeException(
                "Buffer size reached {}".format(self._limit))
        self._buffer.write(buf)

    def __len__(self):
        return len(self.getvalue())
