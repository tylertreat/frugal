import mock
from threading import Thread
import time
import unittest

from thrift.protocol.TJSONProtocol import TJSONProtocolFactory
from thrift.transport.TSocket import TSocket
from thrift.transport.TSocket import TServerSocket

from frugal.protocol import FProtocolFactory
from frugal.server import FSimpleServer


class TestSimpleServer(unittest.TestCase):
    def test_it_works(self):
        processor_factory = mock.Mock()
        mock_processor = mock.Mock()
        processor_factory.get_processor.return_value = mock_processor
        proto_factory = FProtocolFactory(TJSONProtocolFactory())
        server_trans = TServerSocket(host='localhost', port=5536)
        server = FSimpleServer(processor_factory, server_trans, proto_factory)

        thread = Thread(target=lambda: server.serve())
        thread.start()
        time.sleep(0.1)

        transport = TSocket(host='localhost', port=5536)
        transport.open()
        transport.write(bytearray([0, 0, 0, 3, 1, 2, 3]))
        transport.flush()
        time.sleep(0.1)

        server.stop()
        processor_factory.get_processor.assert_called_once_with(mock.ANY)
        mock_processor.process.assert_called_with(mock.ANY, mock.ANY)
