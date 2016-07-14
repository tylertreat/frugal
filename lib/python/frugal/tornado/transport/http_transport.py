import base64
import httplib
from io import BytesIO
import logging
import struct

from thrift.transport.TTransport import TTransportException
from tornado import gen
from tornado.httpclient import HTTPClient
from tornado.httpclient import HTTPError
from tornado.httpclient import HTTPRequest

from frugal.tornado.transport import TTornadoTransportBase

logger = logging.getLogger(__name__)


class FHttpTransport(TTornadoTransportBase):
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
        super(FHttpTransport, self).__init__(max_message_size=request_capacity)
        self._url = url
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

        self._execute = None

    @gen.coroutine
    def isOpen(self):
        """True if the transport is open, False otherwise."""
        with (yield self._open_lock.acquire()):
            # Tornado requires we raise a special exception to return a value.
            raise gen.Return(self._http is not None)

    @gen.coroutine
    def open(self):
        """Opens the transport."""
        with (yield self._open_lock.acquire()):
            if self._http is None:
                self._http = HTTPClient()

    @gen.coroutine
    def close(self):
        """Closes the transport."""
        with (yield self._open_lock.acquire()):
            if self._http is not None:
                self._http.close()
                self._http = None

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

        if not self._execute:
            raise TTransportException(type=TTransportException.UNKNOWN,
                                      message='callback is not set')
        self._execute(decoded)
