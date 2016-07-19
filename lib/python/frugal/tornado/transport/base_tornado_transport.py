from io import BytesIO
import logging
from struct import pack
from threading import Lock

from thrift.transport.TTransport import TTransportException
from tornado import gen, locks

from frugal.exceptions import FMessageSizeException
from frugal.transport import FTransport

logger = logging.getLogger(__name__)


class FTornadoTransportBase(FTransport):
    """ FBaseTransport implements the buffered write data and registry
    interactions shared by all FTransports.
    """

    def __init__(self, max_message_size=1024 * 1024):
        super(FTornadoTransportBase, self).__init__()
        self._max_message_size = max_message_size
        self._wbuf = BytesIO()

        # TODO: Why do we need this lock? This is a threading lock?
        self._lock = Lock()

        # TODO: Remove this with 2.0
        self._execute = None
        self._open_lock = locks.Lock()

    # TODO: Remove with 2.0
    def set_execute_callback(self, execute):
        """Set the message callback execute function

        Args:
            execute: message callback execute function

        @deprecated
        """
        self._execute = execute

    def set_registry(self, registry):
        """Set FRegistry on the transport.  No-Op if already set.
        args:
            registry: FRegistry to set on the transport
        """
        with self._lock:
            if not registry:
                ex = ValueError("registry cannot be null.")
                logger.exception(ex)
                raise ex

            if self._registry:
                return

            self._registry = registry

    def register(self, context, callback):
        with self._lock:
            if not self._registry:
                ex = StandardError("registry cannot be null.")
                logger.exception(ex)
                raise ex

            self._registry.register(context, callback)

    def unregister(self, context):
        with self._lock:
            if not self._registry:
                ex = StandardError("registry cannot be null.")
                logger.exception(ex)
                raise ex

            self._registry.unregister(context)

    def read(self, size):
        raise NotImplementedError("Don't call this.")

    @gen.coroutine
    def write(self, buff):
        """Writes the bytes to a buffer. Throws FMessageSizeException if the
        buffer exceeds limit.

        Args:
            buff: buffer to append to the write buffer.
        """
        if not (yield self.isOpen()):
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "Transport not open!")
            logger.exception(ex)
            raise ex

        size = len(buff) + len(self._wbuf.getvalue())

        if size > self._max_message_size > 0:
            ex = FMessageSizeException("Message exceeds max message size")
            logger.exception(ex)
            raise ex

        self._wbuf.write(buff)

    def get_request_bytes(self):
        frame = self._wbuf.getvalue()
        if len(frame) == 0:
            return None

        frame_length = pack('!I', len(frame))
        self._wbuf = BytesIO()
        return '{0}{1}'.format(frame_length, frame)

    def execute(self, frame):
        self._registry.execute(frame[4:])


class TTornadoTransportBase(FTornadoTransportBase):
    """
    @deprecated Use FTornadoTransportBase instead
    """

    def __init__(self, max_message_size=1024 * 1024):
        super(TTornadoTransportBase, self).__init__(
            max_message_size=max_message_size)
