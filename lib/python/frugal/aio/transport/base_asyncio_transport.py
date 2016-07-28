import asyncio
from io import BytesIO

from thrift.transport.TTransport import TTransportException

from frugal.exceptions import FMessageSizeException
from frugal.transport import FTransport


class FAsyncIOTransportBase(FTransport):
    def __init__(self, max_message_size):
        super().__init__()
        self._max_message_size = max_message_size
        self._wbuf = BytesIO()

    def isOpen(self):
        raise NotImplementedError('You must override this')

    @asyncio.coroutine
    def open(self):
        raise NotImplementedError('You must override this')

    @asyncio.coroutine
    def close(self):
        raise NotImplementedError('You must override this')

    def read(self, size):
        raise NotImplementedError('Do not call this')

    def write(self, buf):
        if not self.isOpen():
            raise TTransportException(TTransportException.NOT_OPEN,
                                      'Transport is not open')

        size = len(buf) + len(self._wbuf.getvalue())

        if size > self._max_message_size > 0:
            raise FMessageSizeException('Message exceeds max message size')
        self._wbuf.write(buf)

    @asyncio.coroutine
    def flush(self):
        raise NotImplementedError('You must override this')
