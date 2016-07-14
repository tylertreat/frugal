import base64
import httplib
from io import BytesIO
import logging
import struct

from thrift.transport.TTransport import TTransportBase
from thrift.transport.TTransport import TTransportException
from tornado import gen
from tornado.httpclient import HTTPClient
from tornado.httpclient import HTTPError
from tornado.httpclient import HTTPRequest

from frugal.exceptions import FMessageSizeException

logger = logging.getLogger(__name__)


class FHttpTransport(TTransportBase):
    def __init__(self, url, request_capacity=0, response_capacity=0):
        """
        Create an HTTP transport.

        :param url: The url to send requests to.
        :type url: str
        :param request_capacity: The maximum size allowed to be written in a
                                 request. Set to 0 for no size restrictions.
        :type request_capacity: int
        :param response_capacity: The maximum size allowed to be read in a
                                  response. Set to 0 for no size restrictions.
        :type response_capacity: int
        """
        self._url = url
        self._request_capacity = request_capacity
        self._http = None
        self._wbuf = BytesIO()

        # create headers
        self._headers = {
            'content-type': 'application/x-frugal',
            'content-transfer-encoding': 'base64',
            'accept': 'application/x-frugal',
        }
        if response_capacity > 0:
            self._headers['x-frugal-payload-limit'] = str(response_capacity)

        self._execute_callback = None

    def set_execute_callback(self, execute_callback):
        """
        Set a callback to be executed with the response body.

        :param execute_callback: The callback to be executed.
        """
        self._execute_callback = execute_callback

    def isOpen(self):
        """True if the transport is open, False otherwise."""
        return self._http is not None

    @gen.coroutine
    def open(self):
        """Opens the transport."""
        if self._http is None:
            self._http = HTTPClient()

    @gen.coroutine
    def close(self):
        """Closes the transport."""
        if self._http is not None:
            self._http.close()
            self._http = None

    def read(self, sz):
        """You should not call read on the HTTP transport"""
        ex = NotImplementedError('read should not be called')
        logger.exception(ex)
        raise ex

    def write(self, buf):
        """
        Write the given data to the transport's buffer.

        :param buf: The data to write to the transport.
        """
        if len(self._wbuf.getvalue()) + len(buf) > self._request_capacity > 0:
            size = len(self._wbuf.getvalue()) + len(buf)
            # clear
            self._wbuf = BytesIO()
            message = 'message exceeded {0} bytes, was {1} bytes'.format(
                    self._request_capacity, size
            )
            raise FMessageSizeException(message=message)

        self._wbuf.write(buf)

    @gen.coroutine
    def flush(self):
        """
        Write the current buffer and execute the set callback with the response.
        """
        frame = self._wbuf.getvalue()
        if len(frame) == 0:
            return

        self._wbuf = BytesIO()
        frame_length = struct.pack('!I', len(frame))
        encoded = base64.b64encode(frame_length + frame)
        request = HTTPRequest(self._url, method='POST', body=encoded,
                              headers=self._headers)

        try:
            response = self._http.fetch(request)
        except HTTPError as e:
            if e.code == httplib.REQUEST_ENTITY_TOO_LARGE:
                raise TTransportException(type=TTransportException.UNKNOWN,
                                          message='response was too large')

            message = 'response errored with code {0} and body {1}'.format(
                    e.code, e.message
            )
            raise TTransportException(type=TTransportException.UNKNOWN,
                                      message=message)

        decoded = base64.b64decode(response.body)

        if len(decoded) < 4:
            raise TTransportException(type=TTransportException.UNKNOWN,
                                      message='invalid frame size')

        if len(decoded) == 4:
            # One-way method, drop response
            return

        self._execute_callback(decoded)
