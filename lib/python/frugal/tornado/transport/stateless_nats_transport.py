from io import BytesIO
import logging
from struct import pack
from threading import Lock

from nats.io.utils import new_inbox
from thrift.transport.TTransport import TTransportBase, TTransportException
from tornado import gen

from frugal.exceptions import FMessageSizeException

logger = logging.getLogger(__name__)

_NATS_MAX_MESSAGE_SIZE = 1024 * 1024
_FRAME_BUFFER_SIZE = 5


class TStatelessNatsTransport(TTransportBase):

    def __init__(self, nats_client, subject, inbox=""):
        self._nats_client = nats_client
        self._subject = subject
        self._inbox = inbox or new_inbox()
        self._is_open = False
        self._open_lock = Lock()
        self._wbuf = BytesIO()

    def set_execute_callback(self, execute):
        self._execute = execute

    def isOpen(self):
        with self._open_lock:
            return self._is_open and self._nats_client.is_connected()

    @gen.coroutine
    def open(self):
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
            self._subject,
            "",
            self._on_message_callback
        )
        self._is_open = True

    @gen.coroutine
    def _on_message_callback(self, msg):
        wrapped = bytearray(msg.data)
        self._execute(wrapped)

    @gen.coroutine
    def close(self):
        if not self._sub_id:
            return

        yield self._nats_client.unsubscribe(self._sub_id)
        self._is_open = False

    def read(self, sz):
        ex = Exception("Don't call this.")
        logger.exception(ex)
        raise ex

    def write(self, buff):
        if not self.isOpen():
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "Nats not connected!")
            logger.exception(ex)
            raise ex

        wbuf_length = len(self._wbuf.getvalue())

        size = len(buff) + wbuf_length

        if size > _NATS_MAX_MESSAGE_SIZE:
            ex = FMessageSizeException("Message exceeds NATS max message size")
            logger.exception(ex)
            raise ex

        self._wbuf.write(buff)

    @gen.coroutine
    def flush(self):
        if not self.isOpen():
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "Nats not connected!")
            logger.exception(ex)
            raise ex

        frame = self._wbuf.getvalue()
        frame_length = pack('!I', len(frame))
        self._wbuf = BytesIO()

        formatted_subject = self._get_formatted_subject()

        yield self._nats_client.publish(formatted_subject, frame_length + frame)

    def _get_formatted_subject(self):
        # TODO: use constant
        return "{}{}".format("frugal.", self._subject)

