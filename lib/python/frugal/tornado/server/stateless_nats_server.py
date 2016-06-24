import logging
import struct

from thrift.Thrift import TException
from tornado import gen

from frugal.server import FServer
from frugal.transport import FTransport
from frugal.tornado.transport import FBoundedMemoryBuffer

logger = logging.getLogger(__name__)

_NATS_PROTOCOL_V0 = 0


class FStatelessNatsTornadoServer(FServer):

    def __init__(self,
                 nats_client,
                 subject,
                 processor,
                 protocol_factory,
                 high_watermark=FTransport.DEFAULT_HIGH_WATERMARK,
                 queue=""):
        """Create a new instance of FStatelessNatsTornadoServer

        Args:
            nats_client: connected instance of nats.io.Client
            subject: subject to listen on
            processor: FProcess
            protocol_factory: FProtocolFactory
        """
        self._nats_client = nats_client
        self._subject = subject
        self._processor = processor
        self._iprot_factory = protocol_factory
        self._oprot_factory = protocol_factory
        self._high_watermark = high_watermark
        self._queue = queue
        self._sub_id = None

    @gen.coroutine
    def serve(self):
        """Subscribe to provided subject and listen on "rpc" queue."""

        self._sub_id = yield self._nats_client.subscribe(
            self._subject,
            self._queue,
            self._on_message_callback
        )

        logger.debug("Frugal server started.")

    @gen.coroutine
    def stop(self):
        """Stop listening for RPC calls."""
        logger.debug("Shutting down Frugal NATS Server.")
        yield self._nats_client.unsubscribe(self._sub_id)

    def set_high_watermark(self, high_watermark):
        """Set the high watermark value for the server

        Args:
            high_watermark: long representing high watermark value
        """
        self._high_watermark = high_watermark

    def get_high_watermark(self):
        # TODO this should be thread safe
        return self._high_watermark

    @gen.coroutine
    def _on_message_callback(self, msg):
        reply_to = msg.reply
        if not reply_to:
            logger.warn("Discarding invalid NATS request (no reply)")
            return

        iprot = self._iprot_factory.get_protocol(
            FBoundedMemoryBuffer(msg.data[4:])
        )
        out_transport = FBoundedMemoryBuffer()
        oprot = self._oprot_factory.get_protocol(out_transport)

        try:
            self._processor.process(iprot, oprot)
        except TException as ex:
            logger.exception(ex)

        if len(out_transport) == 0:
            return

        data = out_transport.getvalue()
        frame_len = len(data)
        buff = bytearray(4)
        struct.pack_into('!I', buff, 0, frame_len + 4)

        yield self._nats_client.publish(reply_to, buff + data)
