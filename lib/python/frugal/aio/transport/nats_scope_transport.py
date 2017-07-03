# Copyright 2017 Workiva
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import inspect

from nats.aio.client import Client
from thrift.transport.TTransport import TTransportException
from thrift.transport.TTransport import TMemoryBuffer

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.exceptions import TTransportExceptionType
from frugal.transport import FPublisherTransport
from frugal.transport import FSubscriberTransport
from frugal.transport import FPublisherTransportFactory
from frugal.transport import FSubscriberTransportFactory


class FNatsPublisherTransportFactory(FPublisherTransportFactory):
    def __init__(self, nats_client: Client):
        self._nats_client = nats_client

    def get_transport(self) -> FPublisherTransport:
        return FNatsPublisherTransport(self._nats_client)


class FNatsPublisherTransport(FPublisherTransport):
    """
    FPublisherTransport is used exclusively for pub/sub scopes. Publishers use
    it to publish to a topic. Nats is used as the underlying bus.
    """
    def __init__(self, nats_client: Client):
        super().__init__(_NATS_MAX_MESSAGE_SIZE)
        self._nats_client = nats_client

    async def open(self):
        if not self._nats_client.is_connected:
            raise TTransportException(TTransportExceptionType.NOT_OPEN,
                                      'Nats is not connected')

    async def close(self):
        if not self.is_open():
            return

        await self._nats_client.flush()

    def is_open(self) -> bool:
        return self._nats_client.is_connected

    async def publish(self, topic: str, data):
        if not self.is_open():
            raise TTransportException(TTransportExceptionType.NOT_OPEN,
                                      'Transport is not connected')
        if self._check_publish_size(data):
            raise TTransportException(
                type=TTransportExceptionType.REQUEST_TOO_LARGE,
                message='Message exceeds max message size'
            )
        await self._nats_client.publish('frugal.{0}'.format(topic), data)


class FNatsSubscriberTransportFactory(FSubscriberTransportFactory):
    def __init__(self, nats_client: Client, queue=''):
        self._nats_client = nats_client
        self._queue = queue

    def get_transport(self) -> FSubscriberTransport:
        return FNatsSubscriberTransport(self._nats_client, self._queue)


class FNatsSubscriberTransport(FSubscriberTransport):
    """
    FSubscriberTransport is used exclusively for pub/sub scopes. Subscribers use
    it to subscribe to a pub/sub topic. Nats is used as the underlying bus.
    """
    def __init__(self, nats_client: Client, queue=''):
        self._nats_client = nats_client
        self._queue = queue
        self._is_subscribed = False
        self._sub_id = None

    async def subscribe(self, topic: str, callback):
        if not self._nats_client.is_connected:
            raise TTransportException(TTransportExceptionType.NOT_OPEN,
                                      'Nats is not connected')
        if self.is_subscribed():
            raise TTransportException(TTransportExceptionType.ALREADY_OPEN,
                                      'Already subscribed to nats topic')

        async def nats_callback(message):
            ret = callback(TMemoryBuffer(message.data[4:]))
            if inspect.iscoroutine(ret):
                ret = await ret
            return ret

        self._sub_id = await self._nats_client.subscribe(
            'frugal.{0}'.format(topic),
            queue=self._queue,
            cb=nats_callback,
        )
        self._is_subscribed = True

    async def unsubscribe(self):
        await self._nats_client.unsubscribe(self._sub_id)
        self._sub_id = None
        self._is_subscribed = False

    def is_subscribed(self) -> bool:
        return self._is_subscribed
