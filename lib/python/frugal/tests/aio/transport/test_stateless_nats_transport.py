import asyncio

import mock
from nats.aio.client import Client
from thrift.transport.TTransport import TTransportException

from frugal.aio.transport import FStatelessNatsAsyncIOTransport
from frugal.exceptions import FExecuteCallbackNotSet
from frugal.tests.aio import utils


class TestTStatelessNatsAsyncIOTransport(utils.AsyncIOTestCase):

    def setUp(self):
        super().setUp()
        self.mock_nats_client = mock.Mock(spec=Client)
        self.subject = 'foo'
        self.inbox = 'bar'
        self.transport = FStatelessNatsAsyncIOTransport(
            self.mock_nats_client,
            self.subject,
            inbox=self.inbox
        )

    @mock.patch('frugal.aio.transport.stateless_nats_transport.new_inbox')
    def test_init(self, mock_new_inbox):
        self.assertEqual(self.mock_nats_client, self.transport._nats_client)
        self.assertEqual(self.subject, self.transport._subject)
        self.assertEqual(self.inbox, self.transport._inbox)

        mock_new_inbox.return_value = 'a new inbox'
        transport = FStatelessNatsAsyncIOTransport(self.mock_nats_client,
                                                   self.subject)
        mock_new_inbox.assert_called_once_with()
        self.assertEqual('a new inbox', transport._inbox)

    @utils.async_runner
    async def test_open_nats_not_connected(self):
        self.mock_nats_client.is_connected = False

        with self.assertRaises(TTransportException) as e:
            await self.transport.open()
            self.assertEqual(TTransportException.NOT_OPEN, e.type)

    @utils.async_runner
    async def test_open_already_open(self):
        self.mock_nats_client.is_connected = True
        self.transport._is_open = True

        with self.assertRaises(TTransportException) as e:
            await self.transport.open()
            self.assertEqual(TTransportException.ALREADY_OPEN, e.type)

    @utils.async_runner
    async def test_open_subscribes(self):
        future = asyncio.Future()
        future.set_result(235)
        self.mock_nats_client.subscribe.return_value = future
        await self.transport.open()

        self.assertEqual(235, self.transport._sub_id)
        self.mock_nats_client.subscribe.assert_called_once_with(
            self.inbox,
            cb=self.transport._on_message_callback
        )
        self.assertTrue(self.transport._is_open)

    # def test_on_message_callback_none(self):
    #     self.transport._callback = None
    #     with self.assertRaises(FExecuteCallbackNotSet):
    #         self.transport._on_message_callback(None)

    def test_on_message_callback(self):
        message = mock.Mock()
        callback = mock.Mock()
        # self.transport.set_execute_callback(callback)
        self.transport.execute = callback
        self.transport._on_message_callback(message)
        callback.assert_called_once_with(message.data)

    @utils.async_runner
    async def test_close_not_subscribed(self):
        self.transport._sub_id = None
        await self.transport.close()
        self.assertFalse(self.mock_nats_client.unsubscribe.called)

    @utils.async_runner
    async def test_close_unsubscribes(self):
        self.transport._is_open = True
        self.transport._sub_id = 235
        future = asyncio.Future()
        future.set_result(None)
        self.mock_nats_client.unsubscribe.return_value = future
        await self.transport.close()

        self.assertIsNone(self.transport._sub_id)
        self.assertFalse(self.transport._is_open)
        self.mock_nats_client.unsubscribe.assert_called_once_with(235)

    @utils.async_runner
    async def test_flush_not_open(self):
        self.transport._is_open = False
        with self.assertRaises(TTransportException) as e:
            await self.transport.flush()
            self.assertEqual(TTransportException.NOT_OPEN, e.type)

    @utils.async_runner
    async def test_flush(self):
        self.transport._is_open = True
        data = bytearray([2, 3, 4, 5, 6, 7])
        self.transport._wbuf.write(data)
        future = asyncio.Future()
        future.set_result(None)
        self.mock_nats_client.publish_request.return_value = future
        await self.transport.flush()

        self.assertEqual(0, len(self.transport._wbuf.getvalue()))
        self.mock_nats_client.publish_request.assert_called_once_with(
            self.subject,
            self.inbox,
            bytearray([0, 0, 0, 6]) + data
        )


















