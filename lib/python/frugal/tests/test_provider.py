import unittest
from mock import patch

from frugal.transport.transport import FTransport
from frugal.protocol import FProtocolFactory
from frugal.provider import FScopeProvider


class TestFScopeProvider(unittest.TestCase):

    @patch('frugal.transport.transport_factory.FPublisherTransportFactory')
    @patch('frugal.transport.transport_factory.FSubscriberTransportFactory')
    @patch('thrift.protocol.TProtocol.TProtocolBase')
    def test_new_provider(self, mock_pub_transport_factory,
                          mock_sub_transport_factory, mock_thrift_protocol):

        pub_transport = FTransport()
        sub_transport = FTransport()
        protocol_factory = FProtocolFactory(None)

        mock_pub_transport_factory.get_transport.return_value = pub_transport
        mock_sub_transport_factory.get_transport.return_value = sub_transport

        provider = FScopeProvider(mock_pub_transport_factory,
                                  mock_sub_transport_factory, protocol_factory)

        trans, prot_factory = provider.new_publisher()

        self.assertEqual(pub_transport, trans)
        self.assertEqual(protocol_factory, prot_factory)

        trans, prot_factory = provider.new_subscriber()
        self.assertEqual(sub_transport, trans)
        self.assertEqual(protocol_factory, prot_factory)
