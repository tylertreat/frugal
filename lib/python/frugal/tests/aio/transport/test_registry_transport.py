import asyncio
import mock

from frugal.aio.transport import FRegistryTransport
from frugal.tests.aio import utils


class TestFRegistryTransport(utils.AsyncIOTestCase):
    def setUp(self):
        super().setUp()
        self.transport = FRegistryTransport(0)
        self.mock_registry = mock.Mock()
        self.mock_context = mock.Mock()
        self.mock_callback = mock.Mock()

    @utils.async_runner
    async def test_register(self):
        self.transport._registry = self.mock_registry
        future = asyncio.Future()
        future.set_result(None)
        self.mock_registry.register.return_value = future
        await self.transport.register(self.mock_context, self.mock_callback)
        self.mock_registry.register.assert_called_once_with(
            self.mock_context, self.mock_callback)

    @utils.async_runner
    async def test_unregister(self):
        # self.transport.set_registry(self.mock_registry)
        self.transport._registry = self.mock_registry
        future = asyncio.Future()
        future.set_result(None)
        self.mock_registry.unregister.return_value = future
        await self.transport.unregister(self.mock_context)
        self.mock_registry.unregister.assert_called_once_with(
            self.mock_context)

    @utils.async_runner
    async def test_execute(self):
        self.transport._registry = self.mock_registry
        future = asyncio.Future()
        future.set_result(None)
        self.mock_registry.execute.return_value = future
        mock_data = mock.Mock()
        await self.transport.execute_frame(mock_data)
        self.mock_registry.execute.assert_called_once_with(mock_data)
