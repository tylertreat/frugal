import base64
import logging
import struct

from thrift.transport.TTransport import TMemoryBuffer

logger = logging.getLogger(__name__)


class _FHttpException(Exception):
    def __init__(self, code, message=None):
        super(_FHttpException, self).__init__(message)
        self.code = code


class _FHttpRequest:
    """
    FrugalHttpRequest stores data from an http request in a generic format.
    """
    def __init__(self, headers=None, body=b''):
        if headers is None:
            headers = {}
        self.headers = headers
        self.body = body


class _FHttpResponse:
    """
    FrugalHttpResponse returns data to be sent in an http response in a generic
    format.
    """
    def __init__(self, status_code=200, headers=None, body=b''):
        self._status_code = status_code
        if headers is None:
            headers = {}
        self._headers = headers
        self._body = body

    @property
    def status_code(self):
        return self._status_code

    @property
    def headers(self):
        return self._headers

    @property
    def body(self):
        return self._body


class _FHttpRequestHandler:
    """
    FHttpRequestHandler provides functionality to process rpcs from http.

    Reading from/writing to the network of a library of choice is left up to
    the caller.
    """
    def __init__(self, processor, protocol_factory):
        """
        Initializes a FHttpRequestHandler.

        Args:
            processor: The processor to use to handle requests.
            protocol_factory: A protocol factory to serialize/deserialize frugal
                              requests.
        """
        self._processor = processor
        self._protocol_factory = protocol_factory

    def _preprocess_http_request(self, request):
        """
        Performs some common preprocessing on an http request.

        Args:
            request: A FrugalHttpRequest object.
        Returns:
            otrans: The output transport, used for getting the output data
            iprot: The input protocol, given to a processor.
            oprot: The output protocol, given to a processor.
            response_limit: The maximum allowed size of the response,
                            zero if unlimited.
        """
        response_limit = int(request.headers.get('x-frugal-payload-limit', 0))
        payload = base64.b64decode(request.body)
        if len(payload) <= 4:
            logger.exception('invalid request frame length {}'.format(
                len(payload)))
            # return _FHttpResponse(status_code=400)
            raise _FHttpException(400)

        itrans = TMemoryBuffer(payload[4:])
        otrans = TMemoryBuffer()
        iprot = self._protocol_factory.get_protocol(itrans)
        oprot = self._protocol_factory.get_protocol(otrans)
        return otrans, iprot, oprot, response_limit

    def _postprocess_http_request(self, otrans, response_limit):
        """
        Performs some common postprocessing to produce an http response.

        Args:
            otrans: The output transport.
            response_limit: The maximum allowed size of the response,
                            zero if unlimited.
        Returns:
            A FrugalHttpResponse to write back to the client.
        """
        output_data = otrans.getvalue()
        if len(output_data) > response_limit > 0:
            logger.exception('response limit exceeded')
            return _FHttpResponse(status_code=413)

        frame_len = struct.pack('!I', len(output_data))
        frame = base64.b64encode(frame_len + output_data)

        headers = {
            'content-type': 'application/x-frugal',
            'content-transfer-encoding': 'base64',
            'content-length': len(frame),
        }
        return _FHttpResponse(headers=headers, body=frame)

    def _handle_processor_exception(self, e):
        """
        Handles an unexepcted exception from a processor.
        TODO this isn't right, but it replicates what other implementations do.
        We should check the exception type and return a potentially different
        error code.

        Args:
            e: The exception.
        Returns:
            A FrugalHttpResponse.
        """
        logger.exception(e)
        return _FHttpResponse(status_code=400)

    def handle_http_request(self, request):
        """
        Handles an http rpc. This must be implemented because of differences
        in processors (synchronous v tornado coroutine v asyncio etc.). It
        is recommended to use _preprocess_http_request and
        _postprocess_http_request to make this easier. If those function are
        used correctly, only calling the processor and handling processor
        errors should be necessary for implementors.

        Args:
            request: A FrugalHttpRequest object.
        Returns:
            a FrugalHttpResponse object.
        """
        raise NotImplementedError()


class _FSynchronousHttpRequestHandler(_FHttpRequestHandler):
    """An http request handler for synchronous processors."""
    def handle_http_request(self, request):
        try:
            otrans, iprot, oprot, response_limit = \
                self._preprocess_http_request(request)
        except _FHttpException as e:
            return _FHttpResponse(status_code=e.code)

        try:
            self._processor.process(iprot, oprot)
        except Exception as e:
            return self._handle_processor_exception(e)

        return self._postprocess_http_request(otrans, response_limit)
