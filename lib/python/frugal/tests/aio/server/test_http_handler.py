import base64
import mock

from aiohttp import test_utils
from aiohttp.web import Application
from thrift.protocol import TBinaryProtocol

from frugal.protocol import FProtocolFactory
from frugal.aio.server import new_http_handler


class TestFHttpHandler(test_utils.AioHTTPTestCase):
    def setUp(self):
        self.processor = mock.Mock()
        super().setUp()

    def get_app(self, loop):
        app = Application(loop=loop)
        prot_factory = FProtocolFactory(
            TBinaryProtocol.TBinaryProtocolFactory())
        app.router.add_route(
                'POST',
                '/frugal',
                new_http_handler(self.processor, prot_factory)
        )
        return app

    @test_utils.unittest_run_loop
    async def test_basic(self):
        request_data = bytearray([2, 3, 4])
        request_frame = bytearray([0, 0, 0, 3]) + request_data
        request_payload = base64.b64encode(request_frame)
        response_data = bytearray([6, 7, 8, 9])
        response_frame = bytearray([0, 0, 0, 4]) + response_data

        async def process_data(_, oprot):
            oprot.get_transport().write(response_data)
        self.processor.process.side_effect = process_data

        response = await self.client.post('/frugal', data=request_payload)
        self.assertEqual(200, response.status)
        self.assertTrue(self.processor.process.called)
        iprot, _ = self.processor.process.call_args[0]
        self.assertEqual(request_data, iprot.get_transport().getvalue())
        received_payload = base64.b64decode(await response.text())

        self.assertEqual(response_frame, received_payload)
        self.assertEqual('application/x-frugal',
                         response.headers['content-type'])
        self.assertEqual('base64',
                         response.headers['content-transfer-encoding'])

    @test_utils.unittest_run_loop
    async def test_response_too_large(self):
        request_data = bytearray([2, 3, 4])
        request_frame = bytearray([0, 0, 0, 3]) + request_data
        request_payload = base64.b64encode(request_frame)
        response_data = bytearray([6, 7, 8, 9, 10, 11])

        async def process_data(_, oprot):
            oprot.get_transport().write(response_data)
        self.processor.process.side_effect = process_data

        headers = {
            'x-frugal-payload-limit': '5',
        }
        response = await self.client.post('/frugal', data=request_payload,
                                          headers=headers)
        self.assertEqual(413, response.status)
        self.assertTrue(self.processor.process.called)
        iprot, _ = self.processor.process.call_args[0]
        self.assertEqual(request_data, iprot.get_transport().getvalue())
