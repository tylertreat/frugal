import unittest
from mock import Mock

from frugal.context import FContext
from frugal.exceptions import FException
from frugal.exceptions import FProtocolException
from frugal.registry import FRegistryImpl


class TestClientRegistry(unittest.TestCase):

    def test_register(self):
        registry = FRegistryImpl()
        context = FContext("fooid")
        context._set_op_id(123)
        callback = Mock()
        registry.register(context, callback)
        self.assertEqual(1, len(registry._handlers))

    def test_register_with_existing_op_id(self):
        registry = FRegistryImpl()
        context = FContext("fooid")
        context._set_op_id(0)
        callback = Mock()

        registry.register(context, callback)
        with self.assertRaises(FException) as cm:
            registry.register(context, callback)

        self.assertEquals("context already registered", str(cm.exception))

    def test_unregister(self):
        registry = FRegistryImpl()
        context = FContext("fooid")
        context._set_op_id(1)
        callback = Mock()
        registry.register(context, callback)
        self.assertEqual(1, len(registry._handlers))
        registry.unregister(context)
        self.assertEqual(0, len(registry._handlers))

    def test_execute_bad_frame(self):
        registry = FRegistryImpl()
        context = FContext("fooid")
        callback = Mock()
        registry.register(context, callback)

        with self.assertRaises(FProtocolException) as cm:
            registry.execute(b"foo")

        self.assertEquals("Invalid frame size: 3", str(cm.exception))

    def test_execute_frame_missing_op_id(self):
        registry = FRegistryImpl()
        registry._next_opid = 10
        context = FContext("fooid")
        callback = Mock()
        registry.register(context, callback)

        frame = bytearray(b'\x00\x00\x00\x00\x00\x80\x01\x00\x02\x00\x00\x00'
                          b'\x08basePing\x00\x00\x00\x00\x00')

        with self.assertRaises(FException) as cm:
            registry.execute(frame)

        self.assertEquals("Frame missing op_id", str(cm.exception))

    def test_execute_unregistered_op_id(self):
        registry = FRegistryImpl()
        registry._next_opid = 10
        context = FContext("fooid")
        callback = Mock()
        registry.register(context, callback)

        frame = bytearray(
            b'\x00\x00\x00\x00\x0e\x00\x00\x00\x05_opid\x00\x00\x00\x011'
            b'\x80\x01\x00\x02\x00\x00\x00\x08basePing\x00\x00\x00\x00\x00'
        )

        registry.execute(frame)
        assert not callback.called

    def test_execute(self):
        registry = FRegistryImpl()
        registry._next_opid = 0
        context = FContext("fooid")
        callback = Mock()
        registry.register(context, callback)

        frame = bytearray(
            b'\x00\x00\x00\x00\x0e\x00\x00\x00\x05_opid\x00\x00\x00\x011'
            b'\x80\x01\x00\x02\x00\x00\x00\x08basePing\x00\x00\x00\x00\x00'
        )

        registry.execute(frame)

        self.assertEqual(frame, callback.call_args_list[0][0][0].getvalue())
