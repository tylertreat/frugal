import mock

from tornado import concurrent
from tornado.testing import gen_test, AsyncTestCase
from thrift.transport.TTransport import TTransportException

from frugal.tornado.transport import TStatelessNatsTransport


class TestTNatsStatelessTransport(AsyncTestCase):

    def setUp(self):
        self.mock_nats_client = mock.Mock()
        self.subject = "foo"
        self.inbox = "new_inbox"
        super(TestTNatsStatelessTransport, self).setUp()

        self.transport = TStatelessNatsTransport(self.mock_nats_client,
                                                 self.subject,
                                                 self.inbox)

    @mock.patch('frugal.tornado.transport.stateless_nats_transport.new_inbox')
    def test_init(self, mock_new_inbox):
        self.assertEquals(self.mock_nats_client, self.transport._nats_client)
        self.assertEquals(self.subject, self.transport._subject)
        self.assertEquals(self.inbox, self.transport._inbox)

        mock_new_inbox.return_value = "asdf"

        transport = TStatelessNatsTransport(self.mock_nats_client,
                                            self.subject)

        mock_new_inbox.assert_called_with()
        self.assertEquals("asdf", transport._inbox)

    @gen_test
    def test_open_throws_nats_not_connected_exception(self):
        self.mock_nats_client.is_connected.return_value = False

        try:
            yield self.transport.open()
            self.fail()
        except TTransportException as e:
            self.assertEqual(TTransportException.NOT_OPEN, e.type)
            self.assertEqual("NATS not connected.", e.message)

    @gen_test
    def test_open_throws_transport_already_open_exception(self):
        self.mock_nats_client.is_connected.return_value = True
        self.transport._is_open = True

        try:
            yield self.transport.open()
            self.fail()
        except TTransportException as e:
            self.assertEqual(TTransportException.ALREADY_OPEN, e.type)
            self.assertEqual("NATS transport already open", e.message)

    @gen_test
    def test_open(self):
        f = concurrent.Future()
        f.set_result(1)
        self.mock_nats_client.subscribe.return_value = f

        yield self.transport.open()

        self.assertEquals(1, self.transport._sub_id)

    @gen_test
    def test_close(self):
        self.transport._sub_id = 1
        f = concurrent.Future()
        f.set_result(None)
        self.mock_nats_client.unsubscribe.return_value = f

        yield self.transport.close()

        self.mock_nats_client.unsubscribe.assert_called_with(
            self.transport._sub_id)

        self.assertFalse(self.transport._is_open)
