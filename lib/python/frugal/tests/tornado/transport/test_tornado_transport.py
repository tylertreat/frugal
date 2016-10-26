import mock

from tornado.concurrent import Future
from tornado.testing import gen_test, AsyncTestCase

from frugal.tornado.transport import FTornadoTransport


class TestFTornadoTranpsort(AsyncTestCase):

    def setUp(self):
        self.transport = FTornadoTransport()

        super(TestFTornadoTranpsort, self).setUp()

    def test_default_values(self):
        self.assertEquals(1024*1024, self.transport._max_message_size)

    @gen_test
    def test_register_calls_registry_register(self):
        mock_registry = mock.Mock()
        register_future = Future()
        register_future.set_result(None)
        mock_registry.register.return_value = register_future
        self.transport._registry = mock_registry

        yield self.transport.register("ctx", "callback")

        mock_registry.register.assert_called_with("ctx", "callback")

    @gen_test
    def test_unregister_calls_registry_unregister(self):
        mock_registry = mock.Mock()
        unregister_future = Future()
        unregister_future.set_result(None)
        mock_registry.unregister.return_value = unregister_future
        self.transport._registry = mock_registry

        yield self.transport.unregister("ctx")

        mock_registry.unregister.assert_called_with("ctx")

    def test_is_open_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError) as cm:
            self.transport.is_open()

        self.assertEquals("You must override this.", cm.exception.message)

    @gen_test
    def test_open_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError) as cm:
            yield self.transport.open()

        self.assertEquals("You must override this.", cm.exception.message)

    @gen_test
    def test_close_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError) as cm:
            yield self.transport.close()

        self.assertEquals("You must override this.", cm.exception.message)

    @gen_test
    def test_send_raises_not_implemented_error(self):
        with self.assertRaises(NotImplementedError) as cm:
            yield self.transport.send([])

        self.assertEquals("You must override this.", cm.exception.message)

    def test_execute_frame_calls_registry_execute_without_frame_size(self):
        frame = 'asdfg'

        mock_registry = mock.Mock()
        self.transport._registry = mock_registry

        self.transport.execute_frame(frame)

        mock_registry.execute.assert_called_with('g')
