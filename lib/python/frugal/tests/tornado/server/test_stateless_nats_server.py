from io import BytesIO
import mock
import struct
from tornado import concurrent
from tornado.testing import gen_test, AsyncTestCase

from frugal.tornado.server.stateless_nats_server import FStatelessNatsTornadoServer


class TestFStatelessNatsTornadoServer(AsyncTestCase):

    def setUp(self):
        super(TestFStatelessNatsTornadoServer, self).setUp()

        self.subject = "foo"
        self.mock_nats_client = mock.Mock()
        self.mock_processor = mock.Mock()
        self.mock_transport_factory = mock.Mock()
        self.mock_prot_factory = mock.Mock()

        self.server = FStatelessNatsTornadoServer(
            self.mock_nats_client,
            self.subject,
            self.mock_processor,
            self.mock_prot_factory
        )
        self.server._iprot_factory = mock.Mock()
        self.server._oprot_factory = mock.Mock()

    @gen_test
    def test_serve(self):
        f = concurrent.Future()
        f.set_result(123)
        self.mock_nats_client.subscribe.return_value = f

        yield self.server.serve()

        self.assertEquals(123, self.server._sub_id)

    @gen_test
    def test_stop(self):
        self.server._sub_id = 123
        f = concurrent.Future()
        f.set_result(None)
        self.mock_nats_client.unsubscribe.return_value = f

        yield self.server.stop()

        self.mock_nats_client.unsubscribe.assert_called_with(123)

    @gen_test
    def test_on_message_callback(self):
        iprot = BytesIO()
        oprot = BytesIO()
        self.server._iprot_factory.get_protocol.return_value = iprot
        self.server._oprot_factory.get_protocol.return_value = oprot

        msg = TestMsg(subject="foo", reply="inbox", data=b'asdf')

        yield self.server._on_message_callback(msg)

        frame_len = len(msg.data)
        buff = bytearray(4)
        struct.pack_into('!I', buff, 0, frame_len + 4)

        self.server._processor.process.assert_called_with(iprot, oprot)


class TestMsg(object):
    def __init__(self, subject='', reply='', data=b'', sid=0,):
        self.subject = subject
        self.reply = reply
        self.data = data
        self.sid = sid
