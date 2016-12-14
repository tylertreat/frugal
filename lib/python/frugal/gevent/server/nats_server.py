import logging
import struct

import gevent
from thrift.Thrift import TApplicationException
from thrift.transport.TTransport import TMemoryBuffer

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.server import FServer
from frugal.transport import TMemoryOutputBuffer

logger = logging.getLogger(__name__)


class FNatsGeventServer(FServer):
    """An implementation of FServer which uses NATS as the underlying transport.
    Clients must connect with the FNatsTransport"""

    def __init__(self, nats_client, subjects, processor,
                 protocol_factory, queue=""):
        """Create a new instance of FStatelessNatsTornadoServer

        Args:
            nats_client: connected instance of gnats.Client
            subject: subject to listen on
            processor: FProcess
            protocol_factory: FProtocolFactory
        """
        self._nats_client = nats_client
        self._subjects = [subjects] if isinstance(subjects, basestring) \
            else subjects
        self._processor = processor
        self._iprot_factory = protocol_factory
        self._oprot_factory = protocol_factory
        self._queue = queue
        self._subs = []

    def serve(self):
        """Subscribe to provided subject and listen on provided queue"""
        queue = self._queue
        cb = self._on_message_callback

        self._subs = [
            self._nats_client.subscribe(
                subject,
                queue=queue,
                cb=cb
            ) for subject in self._subjects
        ]

        logger.info("Frugal server running...")
        while True:
            gevent.sleep(0)

    def stop(self):
        """Unsubscribe from server subject"""
        logger.debug("Frugal server stopping...")
        for sid in self._sub_ids:
            self._nats_client.unsubscribe(sid)

    def _on_message_callback(self, msg):
        """Process and respond to server request on server subject

        Args:
            msg: request message published to server subject
        """
        reply_to = msg.reply
        if not reply_to:
            logger.warn("Discarding invalid NATS request (no reply)")
            return

        frame_size = struct.unpack('!I', msg.data[:4])[0]
        if frame_size > _NATS_MAX_MESSAGE_SIZE - 4:
            logger.warning("Invalid frame size, dropping message.")
            return

        # Read and process frame (exclude first 4 bytes which
        # represent frame size).
        iprot = self._iprot_factory.get_protocol(
            TMemoryBuffer(msg.data[4:])
        )
        otrans = TMemoryOutputBuffer(_NATS_MAX_MESSAGE_SIZE)
        oprot = self._oprot_factory.get_protocol(otrans)

        try:
            self._processor.process(iprot, oprot)
        except TApplicationException:
            # Continue so the exception is sent to the client
            pass
        except Exception:
            return

        if len(otrans) == 4:
            return

        self._nats_client.publish(reply_to, otrans.getvalue())
