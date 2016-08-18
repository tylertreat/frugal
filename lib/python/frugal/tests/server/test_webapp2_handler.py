import base64
import unittest

import mock
from thrift.protocol import TBinaryProtocol
import webapp2
import webtest

from frugal.protocol import FProtocolFactory
from frugal.server.webapp2_handler import new_webapp2_handler


class FWebapp2HttpHandlerTest(unittest.TestCase):
    def setUp(self):
        self.mock_processor = mock.Mock()
        prot_factory = FProtocolFactory(
                TBinaryProtocol.TBinaryProtocolFactory())
        app = webapp2.WSGIApplication([('/frugal', new_webapp2_handler(
                self.mock_processor, prot_factory))])
        self.test_app = webtest.TestApp(app)

    def test_basic(self):
        request_data = bytearray([2, 3, 4])
        request_frame = bytearray([0, 0, 0, 3]) + request_data
        request_payload = base64.b64encode(request_frame)
        response_data = bytearray([6, 7, 8, 9])
        response_frame = bytearray([0, 0, 0, 4]) + response_data

        def process_data(_, oprot):
            oprot.get_transport().write(response_data)
        self.mock_processor.process.side_effect = process_data

        response = self.test_app.post('/frugal', params=request_payload)
        self.assertEqual(200, response.status_int)
        self.assertTrue(self.mock_processor.process.called)
        iprot, _ = self.mock_processor.process.call_args[0]
        self.assertEqual(request_data, iprot.get_transport().getvalue())

        expected_response_payload = base64.b64encode(response_frame)

        self.assertEqual(expected_response_payload, response.normal_body)
        self.assertEqual('application/x-frugal',
                         response.headers['content-type'])
        self.assertEqual('base64',
                         response.headers['content-transfer-encoding'])

    def test_response_too_large(self):
        request_data = bytearray([2, 3, 4])
        request_frame = bytearray([0, 0, 0, 3]) + request_data
        request_payload = base64.b64encode(request_frame)
        response_data = bytearray([6, 7, 8, 9, 10, 11])

        def process_data(_, oprot):
            oprot.get_transport().write(response_data)
        self.mock_processor.process.side_effect = process_data

        headers = {
            'x-frugal-payload-limit': '5',
        }
        response = self.test_app.post('/frugal', params=request_payload,
                                      headers=headers, status='*')
        self.assertEqual(413, response.status_int)
        self.assertTrue(self.mock_processor.process.called)
        iprot, _ = self.mock_processor.process.call_args[0]
        self.assertEqual(request_data, iprot.get_transport().getvalue())

# class AppTest(unittest.TestCase):
#     def setUp(self):
#         # Create a WSGI application.
#         app = webapp2.WSGIApplication([('/', HelloWorldHandler)])
#         self.testapp = webtest.TestApp(app)
#
#     # Test the handler.
#     def testHelloWorldHandler(self):
#         response = self.testapp.get('/')
#         self.assertEqual(response.status_int, 200)
#         self.assertEqual(response.normal_body, 'Hello World!')
#         self.assertEqual(response.content_type, 'text/plain')
