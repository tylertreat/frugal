from frugal.exceptions import FMessageSizeException
from thrift.transport.TTransport import TMemoryBuffer

_NATS_MAX_MESSAGE_SIZE = 1024 * 1024


class FBoundedMemoryBuffer(TMemoryBuffer, object):
    """An implementation of TMemoryBuffer using a bounded memory buffer.
    Writes which cause the buffer to exceed its size throw an FMessageSizeException"""

    def __init__(self, value=None):
        super(FBoundedMemoryBuffer, self).__init__(value)

    def write(self, buf):
        """Bounded write to buffer"""
        if len(self) + len(buf) > _NATS_MAX_MESSAGE_SIZE:
            self._buffer = TMemoryBuffer()
            raise FMessageSizeException("Buffer size reached {}".format(_NATS_MAX_MESSAGE_SIZE))
        self._buffer.write(buf)

    def __len__(self):
        return len(self.getvalue())
