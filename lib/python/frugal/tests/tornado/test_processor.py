from mock import Mock
import unittest

from thrift.protocol.TBinaryProtocol import TBinaryProtocolFactory
from thrift.transport.TTransport import TMemoryBuffer

from frugal.tornado.processor import FBaseProcessor
from frugal.exceptions import FException
from frugal.protocol import FProtocolFactory
from frugal.transport import TMemoryOutputBuffer


class TestFBaseProcessor(unittest.TestCase):

    def test_process_processor_exception(self):
        processor = FBaseProcessor()
        proc = Mock()
        e = FException('foo bar exception')
        proc.process.side_effect = e
        processor.add_to_processor_map("basePing", proc)
        frame = bytearray(
            b'\x00\x00\x00\x00\x0e\x00\x00\x00\x05_opid\x00\x00\x00\x011'
            b'\x80\x01\x00\x02\x00\x00\x00\x08basePing\x00\x00\x00\x00\x00'
        )
        itrans = TMemoryBuffer(value=frame)
        iprot = FProtocolFactory(TBinaryProtocolFactory()).get_protocol(itrans)
        oprot = Mock()

        yield processor.process(iprot, oprot)

        self.assertEqual(e, cm.exception)

    def test_process_missing_function(self):
        processor = FBaseProcessor()
        frame = bytearray(
            b'\x00\x00\x00\x00\x0e\x00\x00\x00\x05_opid\x00\x00\x00\x011'
            b'\x80\x01\x00\x02\x00\x00\x00\x08basePing\x00\x00\x00\x00\x00'
        )
        itrans = TMemoryBuffer(value=frame)
        iprot = FProtocolFactory(TBinaryProtocolFactory()).get_protocol(itrans)
        otrans = TMemoryOutputBuffer(100)
        oprot = FProtocolFactory(TBinaryProtocolFactory()).get_protocol(otrans)

        yield processor.process(iprot, oprot)

        expected_response = bytearray(
            b'\x00\x00\x00\x50\x00\x00\x00\x00\x0e\x00\x00\x00\x05_opid\x00\x00'
            b'\x00\x011\x80\x01\x00\x03\x00\x00\x00\x08basePing\x00\x00\x00\x00'
            b'\x0b\x00\x01\x00\x00\x00\x1aUnknown function: basePing\x08\x00'
            b'\x02\x00\x00\x00\x01\x00'
        )
        self.assertEqual(otrans.getvalue(), expected_response)

    def test_process(self):
        processor = FBaseProcessor()
        proc = Mock()
        future = Future()
        future.set_result(None)
        proc.process.return_value = future
        processor.add_to_processor_map("basePing", proc)
        frame = bytearray(
            b'\x00\x00\x00\x00\x0e\x00\x00\x00\x05_opid\x00\x00\x00\x011'
            b'\x80\x01\x00\x02\x00\x00\x00\x08basePing\x00\x00\x00\x00\x00'
        )
        itrans = TMemoryBuffer(value=frame)
        iprot = FProtocolFactory(TBinaryProtocolFactory()).get_protocol(itrans)
        oprot = Mock()
        yield processor.process(iprot, oprot)
        assert(proc.process.call_args)
        args, _ = proc.process.call_args
        assert(args[0]._get_op_id() == 1)
        assert(args[1] == iprot)
        assert(args[2] == oprot)
