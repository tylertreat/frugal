import asyncio
from io import BytesIO
import struct

from nats.aio.client import Client
from nats.aio.utils import new_inbox
from thrift.transport.TTransport import TTransportException

from frugal import NATS_MAX_MESSAGE_SIZE
from frugal.aio.transport import FAsyncIORegistryTransport


class FStatelessNatsAsyncIOTransport(FAsyncIORegistryTransport):
    def __init__(
            self,
            nats_client: Client,
            subject: str,
            inbox=''
    ):
        super().__init__(NATS_MAX_MESSAGE_SIZE)
        self._nats_client = nats_client
        self._subject = subject
        self._inbox = inbox or new_inbox()
        self._is_open = False
        self._sub_id = None

    def isOpen(self):
        return self._is_open and self._nats_client.is_connected

    async def open(self):
        if not self._nats_client.is_connected:
            raise TTransportException(TTransportException.NOT_OPEN,
                                      'Nats not connected')

        if self.isOpen():
            raise TTransportException(TTransportException.ALREADY_OPEN,
                                      'Transport is already open')

        self._sub_id = await self._nats_client.subscribe(
            self._inbox,
            cb=self._on_message_callback
        )
        self._is_open = True

    def _on_message_callback(self, message):
        self.execute(message.data)

    async def close(self):
        if not self._sub_id:
            return

        await self._nats_client.unsubscribe(self._sub_id)
        self._is_open = False
        self._sub_id = None

    async def flush(self):
        print('flushing')
        if not self._is_open:
            raise TTransportException(TTransportException.NOT_OPEN,
                                      'Transport is not open')

        data = self._wbuf.getvalue()
        data_length = struct.pack('!I', len(data))
        self._wbuf = BytesIO()
        await self._nats_client.publish_request(
                self._subject,
                self._inbox,
                data_length + data
        )
