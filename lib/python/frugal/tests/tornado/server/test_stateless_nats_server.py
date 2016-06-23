import mock
from tornado import concurrent
from tornado.testing import gen_test, AsyncTestCase

from frugal.tornado.server.stateless_nats_server import FStatelessNatsTornadoServer

_NATS_PROTOCOL_V0 = 0


class TestFStatelessNatsTornadoServer(AsyncTestCase):

    def setUp(self):
        patcher = mock.patch('frugal.tornado.server.stateless_nats_server.new_inbox')
        self.mock_new_inbox = patcher.start()
        self.addCleanup(patcher.stop)

        super(TestFStatelessNatsTornadoServer, self).setUp()
        self.mock_new_inbox.return_value = "new_inbox"

        self.subject = "foo"
        self.mock_nats_client = mock.Mock()
        self.mock_processor = mock.Mock()
        self.mock_transport_factory = mock.Mock()
        self.mock_prot_factory = mock.Mock()

        self.server = FStatelessNatsTornadoServer(
            self.mock_nats_client,
            self.subject,
            self.mock_processor,
            self.mock_transport_factory,
            self.mock_prot_factory
        )

        self.mock_transport = mock.Mock()

    @gen_test
    def test_serve(self):
        f = concurrent.Future()
        f.set_result(123)
        self.mock_nats_client.subscribe.return_value = f

        yield self.server.serve()

        self.assertEquals(123, self.server._sid)

    @gen_test
    def test_stop(self):
        yield self.server.stop()

    def test_set_and_get_high_watermark(self):
        self.server.set_high_watermark(1234)

        self.assertEquals(1234, self.server.get_high_watermark())

    @mock.patch('frugal.tornado.server.nats_server.TNatsServiceTransport')
    @gen_test
    def test_on_message_callback(self, mock_server_constructor):
        mock_server_transport = mock.Mock()

        mock_server_constructor.Server.return_value = mock_server_transport

        f = concurrent.Future()
        f.set_result(None)
        self.mock_transport.open.return_value = f

        self.mock_transport_factory.get_transport.return_value = self.mock_transport
        self.mock_nats_client.publish_request.return_value = f

        msg = TestMsg()

        yield self.server._on_message_callback(msg)
        #TODO test something

    @mock.patch('frugal.tornado.server.nats_server.new_inbox')
    def test_new_frugal_inbox(self, mock_new_inbox):
        mock_new_inbox.return_value = "new_inbox"
        prefix = "frugal._INBOX.d138b9369fa35386624d6ad97"

        result = self.server._new_frugal_inbox(prefix)

        self.assertEquals("frugal._INBOX.new_inbox", result)

        #TODO how to test close?

        #TODO how to test start?


class TestMsg(object):
    def __init__(self, subject='', reply='', data=b'', sid=0,):
        self.subject = subject
        self.reply   = reply
        self.data    = data
        self.sid     = sid
