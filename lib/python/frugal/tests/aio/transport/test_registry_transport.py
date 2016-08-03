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

    def test_set_registry_none(self):
        print('running')
        with self.assertRaises(ValueError):
            self.transport.set_registry(None)

    def test_set_registry_already_set(self):
        self.transport.set_registry(self.mock_registry)
        new_mock_registry = mock.Mock()
        self.transport.set_registry(new_mock_registry)
        self.assertEqual(self.mock_registry, self.transport._registry)

    def test_set_registry(self):
        self.transport.set_registry(self.mock_registry)
        self.assertEqual(self.mock_registry, self.transport._registry)

    def test_register_not_set(self):
        with self.assertRaises(ValueError):
            self.transport.register(self.mock_context, self.mock_callback)

    def test_register(self):
        self.transport.set_registry(self.mock_registry)
        self.transport.register(self.mock_context, self.mock_callback)
        self.mock_registry.register.assert_called_once_with(
                self.mock_context, self.mock_callback)

    def test_unregister_not_set(self):
        with self.assertRaises(ValueError):
            self.transport.unregister(self.mock_context)

    def test_unregister(self):
        self.transport.set_registry(self.mock_registry)
        self.transport.unregister(self.mock_context)
        self.mock_registry.unregister.assert_called_once_with(self.mock_context)

    def test_execute_not_set(self):
        with self.assertRaises(ValueError):
            self.transport.execute_frame([])

    def test_execute(self):
        self.transport.set_registry(self.mock_registry)
        mock_data = mock.Mock()
        self.transport.execute_frame(mock_data)
        self.mock_registry.execute.assert_called_once_with(mock_data)
