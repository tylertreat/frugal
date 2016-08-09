from base64 import b64encode
from struct import pack_into
import sys
import unittest

from mock import Mock
from mock import patch
from thrift.transport.TTransport import TTransportException

# Need to add App Engine SDK to path.
sys.path.insert(1, '/usr/local/google_appengine')

from frugal.transport.urlfetch_transport import FUrlfetchTransport


@patch('frugal.transport.urlfetch_transport.urlfetch')
class TestFUrlfetchTransport(unittest.TestCase):

    def test_request(self, mock_urlfetch):
        url = 'http://localhost:8080/frugal'
        headers = {'foo': 'bar'}
        resp = Mock(status=200)
        response = 'response'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(response))
        resp_body = b64encode(buff + response)
        resp = Mock(status_code=200, content=resp_body)
        mock_urlfetch.fetch.return_value = resp

        def get_headers():
            return {'baz': 'qux'}

        tr = FUrlfetchTransport(url, headers=headers, get_headers=get_headers)
        deadline = 5
        tr.set_timeout(deadline*1000)

        tr.open()
        self.assertTrue(tr.isOpen())

        data = 'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)
        tr.flush()

        mock_urlfetch.fetch.assert_called_once_with(
            url, method=mock_urlfetch.POST, payload=encoded_frame,
            validate_certificate=False, deadline=deadline,
            headers={'foo': 'bar', 'baz': 'qux', 'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/FHttpTransport'})

        resp = tr.read(len(response))
        self.assertEqual(response, resp)

        tr.close()
        self.assertTrue(tr.isOpen())  # open/close are no-ops

    def test_request_https(self, mock_urlfetch):
        url = 'https://localhost:8080/frugal'
        resp = Mock(status=200)
        response = 'response'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(response))
        resp_body = b64encode(buff + response)
        resp = Mock(status_code=200, content=resp_body)
        mock_urlfetch.fetch.return_value = resp

        tr = FUrlfetchTransport(url)

        data = 'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)
        tr.flush()

        mock_urlfetch.fetch.assert_called_once_with(
            url, method=mock_urlfetch.POST, payload=encoded_frame,
            validate_certificate=True, deadline=None,
            headers={'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/FHttpTransport'})

        resp = tr.read(len(response))
        self.assertEqual(response, resp)

    def test_flush_no_body(self, mock_urlfetch):
        url = 'http://localhost:8080/frugal'

        tr = FUrlfetchTransport(url)
        tr.flush()

        self.assertFalse(mock_urlfetch.fetch.called)

    def test_flush_bad_response(self, mock_urlfetch):
        url = 'http://localhost:8080/frugal'
        resp = Mock(status_code=500)
        mock_urlfetch.fetch.return_value = resp

        tr = FUrlfetchTransport(url)

        data = 'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)

        with self.assertRaises(TTransportException):
            tr.flush()

        mock_urlfetch.fetch.assert_called_once_with(
            url, method=mock_urlfetch.POST, payload=encoded_frame,
            validate_certificate=False, deadline=None,
            headers={'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/FHttpTransport'})

    def test_flush_bad_oneway_response(self, mock_urlfetch):
        url = 'http://localhost:8080/frugal'
        buff = bytearray(4)
        pack_into('!I', buff, 0, 10)
        resp_body = b64encode(buff)
        resp = Mock(status_code=200, content=resp_body)
        mock_urlfetch.fetch.return_value = resp

        tr = FUrlfetchTransport(url)

        data = 'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)

        with self.assertRaises(TTransportException):
            tr.flush()

        mock_urlfetch.fetch.assert_called_once_with(
            url, method=mock_urlfetch.POST, payload=encoded_frame,
            validate_certificate=False, deadline=None,
            headers={'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/FHttpTransport'})

    def test_flush_oneway(self, mock_urlfetch):
        url = 'http://localhost:8080/frugal'
        buff = bytearray(4)
        pack_into('!I', buff, 0, 0)
        resp_body = b64encode(buff)
        resp = Mock(status_code=200, content=resp_body)
        mock_urlfetch.fetch.return_value = resp

        tr = FUrlfetchTransport(url)

        data = 'helloworld'
        buff = bytearray(4)
        pack_into('!I', buff, 0, len(data))
        encoded_frame = b64encode(buff + data)

        tr.write(data)
        tr.flush()

        mock_urlfetch.fetch.assert_called_once_with(
            url, method=mock_urlfetch.POST, payload=encoded_frame,
            validate_certificate=False, deadline=None,
            headers={'Content-Length': '20',
                     'Content-Type': 'application/x-frugal',
                     'Content-Transfer-Encoding': 'base64',
                     'User-Agent': 'Python/FHttpTransport'})

        resp = tr.read(10)
        self.assertEqual('', resp)
