import unittest

from frugal.exceptions import FMessageSizeException
from frugal.transport import TMemoryOutputBuffer


class TestTOutputMemoryBuffer(unittest.TestCase):

    def setUp(self):
        super(TestTOutputMemoryBuffer, self).setUp()

        self.buffer = TMemoryOutputBuffer(10)

    def test_write(self):
        self.buffer.write(b'foo')
        self.assertSequenceEqual(b'\x00\x00\x00\x03foo', self.buffer.getvalue())

    def test_write_size_exception(self):
        self.assertEqual(4, len(self.buffer))
        self.buffer.write(bytearray(6))
        self.assertEqual(10, len(self.buffer))
        with self.assertRaises(FMessageSizeException):
            self.buffer.write(bytearray(1))
        self.assertEqual(4, len(self.buffer))

    def test_len(self):
        self.buffer.write(b'12345')
        self.assertEqual(len(self.buffer), 9)
