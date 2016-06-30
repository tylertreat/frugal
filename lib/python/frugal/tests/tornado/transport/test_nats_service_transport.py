import mock
import struct

from thrift.transport.TTransport import TTransportException
from tornado import concurrent
from tornado.testing import gen_test, AsyncTestCase

from frugal.tornado.transport.nats_service_transport import TNatsServiceTransport


class TestTNatsServiceTransport(AsyncTestCase):

    def setUp(self):
        self.subject = "foo"
        self.timeout = 20000
        self.max_missed_heartbeats = 3
        super(TestTNatsServiceTransport, self).setUp()

        self.mock_nats_client = mock.Mock()

        self.transport = TNatsServiceTransport.Client(
            self.mock_nats_client,
            self.subject,
            self.timeout,
            self.max_missed_heartbeats
        )

    @gen_test
    def test_init(self):
        self.assertEqual(self.subject, self.transport._connection_subject)
        self.assertEqual(self.timeout, self.transport._connection_timeout)

        self.assertFalse(self.transport._is_open)

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
        self.mock_nats_client.is_connected.return_value = True

        f = concurrent.Future()
        f.set_result("handshake response 1234")

    @gen_test
    def test_write_throws_not_open_exception(self):
        self.transport._is_open = False

        try:
            self.transport.write(b'')
            self.fail()
        except TTransportException as e:
            self.assertEqual("Transport not open!", e.message)

    @gen_test
    def test_write_adds_buff_to_write_buffer(self):
        self.mock_nats_client.is_connected.return_value = True
        self.transport._is_open = True

        buff = bytearray(10)
        struct.pack_into('>I', buff, 0, 10)

        self.transport.write(buff)

    @gen_test
    def test_close(self):
        self.transport._sub_id = 1
        self._is_open = True

        f = concurrent.Future()
        f.set_result("")
        self.mock_nats_client.publish_request.return_value = f
        self.mock_nats_client.unsubscribe.return_value = f

        yield self.transport.close()

        self.mock_nats_client.unsubscribe.assert_called()

        self.assertFalse(self.transport._is_open)

    def test_read_throws_exception(self):
        try:
            self.transport.read(2)
            self.fail()
        except Exception as ex:
            self.assertEquals("Don't call this.", ex.message)

    def test_write(self):
        b = bytearray(b'test')
        try:
            self.transport.write(b)
            self.fail()
        except TTransportException as ex:
            self.assertEquals("Transport not open!", ex.message)

        self.mock_nats_client.is_connected.return_value = True
        self.transport._is_open = True

        self.transport.write(b)

        self.assertEquals(b, self.transport._wbuf.getvalue())

    @gen_test
    def test_flush(self):
        self.transport._write_to = "foo"
        b = bytearray('test')
        self.transport._wbuf.write(b)
        frame_length = struct.pack('!I', len(b))

        f = concurrent.Future()
        f.set_result("")
        self.mock_nats_client.publish.return_value = f

        yield self.transport.flush()

        self.mock_nats_client.publish.assert_called_with("foo",
                                                         frame_length + b)
