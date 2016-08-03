from asyncio.locks import Lock
from io import BytesIO

from nats.aio.client import Client
from thrift.transport.TTransport import TTransportException
from thrift.transport.TTransport import TMemoryBuffer

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.aio.transport import FTransportBase
from frugal.exceptions import FException
from frugal.transport import FScopeTransport
from frugal.transport import FScopeTransportFactory


class FNatsScopeTransportFactory(FScopeTransportFactory):
    def __init__(
            self,
            nats_client: Client,
            queue=''
    ):
        self._nats_client = nats_client
        self._queue = queue

    def get_transport(self):
        return FNatsScopeTransport(
                self._nats_client,
                queue=self._queue
        )


class FNatsScopeTransport(FScopeTransport, FTransportBase):
    def __init__(
            self,
            nats_client: Client,
            queue=''
    ):
        FTransportBase.__init__(self, _NATS_MAX_MESSAGE_SIZE)
        self._nats_client = nats_client
        self._queue = queue
        self._subject = ''
        self._topic_lock = Lock()
        self._pull = False
        self._is_open = False
        self._sub_id = None
        self._callback = None

    async def lock_topic(self, topic: str):
        if self._pull:
            raise FException('Subscribers cannot lock topics')

        await self._topic_lock.acquire()
        self._subject = topic

    def unlock_topic(self):
        if self._pull:
            raise FException('Subscribers cannot lock topics')

        self._subject = ''
        self._topic_lock.release()

    async def subscribe(self, topic: str, callback):
        self._pull = True
        self._subject = topic
        await self.open(callback=callback)

    def isOpen(self):
        return self._nats_client.is_connected and self._is_open

    async def open(self, callback=None):
        if not self._nats_client.is_connected:
            raise TTransportException(TTransportException.NOT_OPEN,
                                      'Nats is not connected')

        if self._is_open:
            raise TTransportException(TTransportException.ALREADY_OPEN,
                                      'Nats is already open')

        if not self._pull:
            self.reset_write_buffer()
            self._is_open = True
            return

        if not self._subject:
            raise TTransportException(TTransportException.UNKNOWN,
                                      'Subscriber cannot have an empty subject')

        self._sub_id = await self._nats_client.subscribe(
            'frugal.{0}'.format(self._subject),
            queue=self._queue,
            cb=lambda message: callback(TMemoryBuffer(message.data[4:]))
        )
        self._is_open = True

    async def close(self):
        if not self.isOpen():
            return

        if not self._pull:
            await self._nats_client.flush()
            self._is_open = False
            return

        await self._nats_client.unsubscribe(self._sub_id)
        self._sub_id = None
        self._is_open = False

    async def flush(self):
        if not self.isOpen():
            raise TTransportException(TTransportException.NOT_OPEN,
                                      'Transport is not connected')

        frame = self.get_write_bytes()
        self.reset_write_buffer()

        await self._nats_client.publish(
            'frugal.{0}'.format(self._subject),
            frame
        )
