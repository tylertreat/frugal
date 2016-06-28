from io import BytesIO
import logging
from struct import pack
from threading import Lock

from nats.io.utils import new_inbox
from thrift.transport.TTransport import TTransportBase, TTransportException
from tornado import gen

from frugal.exceptions import FMessageSizeException, FExecuteCallbackNotSet
from frugal.tornado.transport.nats_scope_transport import MAX_MESSAGE_SIZE

logger = logging.getLogger(__name__)


class TStatelessNatsTransport(TTransportBase):
    """TStatelessNatsTransport is an extension of thrift.TTransportBase.
    This is a "stateless" transport in the sense that there is no
    connection with a server. A request is simply published to a subject
    and responses are received on another subject. This assumes
    requests/responses fit within a single NATS message."""

    def __init__(self, nats_client, subject, inbox=""):
        """Create a new instance of FStatelessNatsTornadoServer

        Args:
            nats_client: connected instance of nats.io.Client
            subject: subject to publish to
        """
        self._nats_client = nats_client
        self._subject = subject
        self._inbox = inbox or new_inbox()
        self._is_open = False
        self._open_lock = Lock()
        self._wbuf = BytesIO()
        self._execute = None
        self._sub_id = None

    def set_execute_callback(self, execute):
        """Set the message callback execute function

        Args:
            execute: message callback execute function
        """
        self._execute = execute

    def isOpen(self):
        with self._open_lock:
            return self._is_open and self._nats_client.is_connected()

    @gen.coroutine
    def open(self):
        """Subscribes to the configured inbox subject"""
        if not self._nats_client.is_connected():
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "NATS not connected.")
            logger.error(ex.message)
            raise ex

        elif self.isOpen():
            ex = TTransportException(TTransportException.ALREADY_OPEN,
                                     "NATS transport already open")
            logger.error(ex.message)
            raise ex

        self._sub_id = yield self._nats_client.subscribe(
            self._inbox,
            "",
            self._on_message_callback
        )
        self._is_open = True

    def _on_message_callback(self, msg):
        if not self._execute:
            ex = FExecuteCallbackNotSet("Execute callback not set")
            logger.error(ex.message)
            raise ex

        wrapped = bytearray(msg.data)
        self._execute(wrapped)

    @gen.coroutine
    def close(self):
        """Unsubscribes from the inbox subject"""
        if not self._sub_id:
            return

        yield self._nats_client.unsubscribe(self._sub_id)
        self._is_open = False

    def read(self, sz):
        """Don't call this"""
        ex = NotImplementedError("Don't call this.")
        logger.exception(ex)
        raise ex

    def write(self, buff):
        """Writes the bytes to a buffer. Throws FMessageSizeException if the buffer exceeds 1MB"""
        if not self.isOpen():
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "Nats not connected!")
            logger.exception(ex)
            raise ex

        size = len(buff) + len(self._wbuf.getvalue())

        if size > MAX_MESSAGE_SIZE:
            ex = FMessageSizeException("Message exceeds NATS max message size")
            logger.exception(ex)
            raise ex

        self._wbuf.write(buff)

    @gen.coroutine
    def flush(self):
        """Sends the buffered bytes over NATS"""
        if not self.isOpen():
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "Nats not connected!")
            logger.exception(ex)
            raise ex

        frame = self._wbuf.getvalue()
        frame_length = pack('!I', len(frame))
        self._wbuf = BytesIO()

        yield self._nats_client.publish_request(self._subject,
                                                self._inbox,
                                                frame_length + frame)
