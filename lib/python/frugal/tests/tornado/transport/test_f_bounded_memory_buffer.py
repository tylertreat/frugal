import unittest

from frugal.exceptions import FMessageSizeException
from frugal.tornado.transport.f_bounded_memory_buffer import FBoundedMemoryBuffer

_NATS_PROTOCOL_V0 = 0
_NATS_MAX_MESSAGE_SIZE = 1024 * 1024


class TestFBoundedMemoryBuffer(unittest.TestCase):

    def setUp(self):
        super(TestFBoundedMemoryBuffer, self).setUp()

        self.buffer = FBoundedMemoryBuffer()
        
    def test_write(self):
        self.buffer.write(bytearray("foo"))
        self.assertSequenceEqual(bytearray("foo"), self.buffer.getvalue())

    def test_write_size_exception(self):
        with self.assertRaises(FMessageSizeException):
            self.buffer.write(bytearray(_NATS_MAX_MESSAGE_SIZE + 1))

    def test_len(self):
        self.buffer.write(bytearray("12345"))
        self.assertEqual(len(self.buffer), 5)
