from thrift.transport.TTransport import TTransportException

from frugal.aio.transport import FTransportBase
from frugal.exceptions import FMessageSizeException
from frugal.tests.aio import utils


class TestFTransportBase(utils.AsyncIOTestCase):
    def setUp(self):
        super().setUp()
        self.transport = FTransportBase(0)
        self.transport.isOpen = lambda: True

    def test_write_transport_not_open(self):
        self.transport.isOpen = lambda: False
        with self.assertRaises(TTransportException) as cm:
            self.transport.write(bytearray([]))
        self.assertEqual(TTransportException.NOT_OPEN, cm.exception.type)

    def test_write(self):
        data = bytearray([1, 2, 3, 4, 5, 6])
        self.transport.write(data[:3])
        self.transport.write(data[3:5])
        self.transport.write(data[5:])
        self.assertEqual(data, self.transport._wbuf.getvalue())

    def test_write_over_limit(self):
        self.transport._max_message_size = 4
        self.transport.write(bytearray([0] * 4))
        with self.assertRaises(FMessageSizeException):
            self.transport.write(bytearray([0]))
