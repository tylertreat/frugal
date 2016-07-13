import logging

from thrift.transport.TTransport import TTransportBase, TTransportException
from tornado import gen, locks

from frugal.exceptions import FMessageSizeException

logger = logging.getLogger(__name__)


class TTornadoTransportBase(TTransportBase, object):

    def __init__(self, max_message_size=1024*1024):
        self._open_lock = locks.Lock()
        self._max_message_size = max_message_size

    def set_execute_callback(self, execute):
        """Set the message callback execute function

        Args:
            execute: message callback execute function
        """
        self._execute = execute

    def read(self, size):
        raise NotImplementedError("Don't call this.")

    @gen.coroutine
    def write(self, buff):
        """Writes the bytes to a buffer.  Throws FMessageSizeException if
        the buffer exceeds limit.

        Args:
            buff: buffer to append to the write buffer.
        """
        if not (yield self.isOpen()):
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "Transport not open!")
            logger.exception(ex)
            raise ex

        size = len(buff) + len(self._wbuf.getvalue())

        if size > self._max_message_size:
            ex = FMessageSizeException("Message exceeds max message size")
            logger.exception(ex)
            raise ex

        self._wbuf.write(buff)

