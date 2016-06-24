import logging
import re

from nats.io.utils import new_inbox
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
        self._input_protocol_factory = protocol_factory
        self._output_protocol_factory = protocol_factory
        self._high_watermark = high_watermark
        self._queue = queue
        self._sid = ""

    @gen.coroutine
    def serve(self):
        """Subscribe to provided subject and listen on "rpc" queue."""

        self._sid = yield self._nats_client.subscribe(
            self._subject,
            self._queue,
            self._on_message_callback
        )

        logger.debug("Frugal server started.")

    @gen.coroutine
    def stop(self):
        """Stop listening for RPC calls."""
        logger.debug("Shutting down Frugal NATS Server.")
        self._nats_client.unsubscribe(self._sid)

    def set_high_watermark(self, high_watermark):
        """Set the high watermark value for the server

        Args:
            high_watermark: long representing high watermark value
        """
        self._high_watermark = high_watermark

    def get_high_watermark(self):
        return self._high_watermark

    def _new_frugal_inbox(self, prefix):
        tokens = re.split('\.', prefix)
        tokens[len(tokens) - 1] = new_inbox()
        inbox = ""
        pre = ""
        for token in tokens:
            inbox += pre + token
            pre = "."
        return inbox

    @gen.coroutine
    def _on_message_callback(self, msg):
        reply_to = msg.reply
        if not reply_to:
            logger.warn("Discarding invalid NATS request (no reply)")
            return

        in_transport = FBoundedMemoryBuffer(msg.data[4:])
        out_transport = FBoundedMemoryBuffer()  # this may just need to be bytearray() or BytesIO()

        try:
            self._processor.process(self._input_protocol_factory.get_protocol(in_transport),
                                    self._output_protocol_factory.get_protocol(out_transport))
        except TException as ex:
            logger.warning("error processing frame: {}".format(ex.message))
            print "got an exception"

        if len(out_transport) == 0:
            return

        response = bytearray(out_transport.getvalue())

        yield self._nats_client.publish(msg.reply, response)
