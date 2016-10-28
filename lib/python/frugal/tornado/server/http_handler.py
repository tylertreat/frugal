import base64
import struct

from thrift.transport.TTransport import TMemoryBuffer
from tornado.web import RequestHandler

from frugal.transport import TMemoryOutputBuffer


class FTornadoHttpHandler(RequestHandler):
    """
    This class implements a Tornado web server request handler to interface
    with a frugal HTTP client.
    """
    def initialize(self, processor, protocol_factory):
        """

        Args:
            processor: The processor to use to handle requests
            protocol_factory: A protocol factory to serialize/deserialize
                              frugal requests
        """
        self._processor = processor
        self._protocol_factory = protocol_factory

    def post(self):
        self.set_header('content-type', 'application/x-frugal')

        # check for response size limit
        response_limit = 0
        if self.request.headers.get('x-frugal-payload-limit') is not None:
            response_limit = int(self.request.headers['x-frugal-payload-limit'])

        # decode payload and process
        payload = base64.b64decode(self.request.body)
        iprot = self._protocol_factory.get_protocol(
            TMemoryBuffer(payload[4:])
        )
        # TODO could be better with this limit
        otrans = TMemoryOutputBuffer(0)
        oprot = self._protocol_factory.get_protocol(otrans)
        # TODO 2.0 this should probably be in a try/catch but we need to decide
        # what to do here
        self._processor.process(iprot, oprot)

        # write back response
        output_data = otrans.getvalue()
        if len(output_data) > response_limit > 0:
            self.send_error(status_code=413)
            return

        output_payload = base64.b64encode(output_data)

        self.set_header('content-transfer-encoding', 'base64')
        self.write(output_payload)
