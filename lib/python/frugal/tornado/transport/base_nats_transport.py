import logging

from thrift.transport.TTransport import TTransportBase, TTransportException

from frugal.tornado.transport.nats_scope_transport import MAX_MESSAGE_SIZE
from frugal.exceptions import FMessageSizeException

logger = logging.getLogger(__name__)


class TNatsTransportBase(TTransportBase, object):

    def __init__(self, nats_client):
        self._nats_client = nats_client

    def isOpen(self):
        return self._is_open and self._nats_client.is_connected()

    def set_execute_callback(self, execute):
        """Set the message callback execute function

        Args:
            execute: message callback execute function
        """
        self._execute = execute

    def read(self, size):
        raise NotImplementedError("Don't call this.")

    def write(self, buff):
        """Writes the bytes to a buffer.
        Throws FMessageSizeException if the buffer exceeds 1MB"""
        if not self.isOpen():
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "Transport not open!")
            logger.exception(ex)
            raise ex

        size = len(buff) + len(self._wbuf.getvalue())

        if size > MAX_MESSAGE_SIZE:
            ex = FMessageSizeException("Message exceeds NATS max message size")
            logger.exception(ex)
            raise ex

        self._wbuf.write(buff)

