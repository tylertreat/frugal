from io import BytesIO
import logging
from struct import pack

from nats.io.utils import new_inbox
from thrift.transport.TTransport import TTransportException
from tornado import gen, locks

from frugal.exceptions import FExecuteCallbackNotSet
from frugal.tornado.transport import TTornadoTransportBase

logger = logging.getLogger(__name__)


class TStatelessNatsTransport(TTornadoTransportBase):
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
        super(TStatelessNatsTransport, self).__init__()
        self._nats_client = nats_client
        self._subject = subject
        self._inbox = inbox or new_inbox()
        self._is_open = False
        self._wbuf = BytesIO()
        self._execute = None
        self._sub_id = None
        self._open_lock = locks.Lock()

    @gen.coroutine
    def isOpen(self):
        with (yield self._open_lock.acquire()):
            # Tornado requires we raise a special exception to return a value.
            raise gen.Return(self._is_open and self._nats_client.is_connected())

    @gen.coroutine
    def open(self):
        """Subscribes to the configured inbox subject"""
        if not self._nats_client.is_connected():
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "NATS not connected.")
            logger.error(ex.message)
            raise ex

        elif (yield self.isOpen()):
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

        self._execute(msg.data)

    @gen.coroutine
    def close(self):
        """Unsubscribes from the inbox subject"""
        if not self._sub_id:
            return

        yield self._nats_client.unsubscribe(self._sub_id)
        with (yield self._open_lock.acquire()):
            self._is_open = False

    @gen.coroutine
    def flush(self):
        """Sends the buffered bytes over NATS"""
        if not (yield self.isOpen()):
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "Nats not connected!")
            logger.exception(ex)
            raise ex

        frame = self._wbuf.getvalue()
        frame_length = pack('!I', len(frame))
        self._wbuf = BytesIO()

        yield self._nats_client.publish_request(self._subject,
                                                self._inbox,
                                                '{0}{1}'.format(frame_length,
                                                                frame))
